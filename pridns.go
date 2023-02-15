package pridns

import (
	"context"
	"fmt"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/google/uuid"
	"github.com/laeni/pri-dns/db"
	myForward "github.com/laeni/pri-dns/forward"
	"github.com/laeni/pri-dns/types"
	"github.com/laeni/pri-dns/util"
	"github.com/miekg/dns"
	"net"
	"sync"
	"time"
)

const (
	storeTypeMySQL = "mysql" // 表示使用 mysql 作为存储介质
	storeTypeEtcd  = "etcd"  // 表示使用 etcd 作为存储介质
	storeTypeRedis = "redis" // 表示使用 redis 作为存储介质
)

var log = clog.NewWithPlugin("pri-dns")

type address struct {
	name string
	ads  []string
}

type PriDns struct {
	Config *types.Config
	Next   plugin.Handler
	Store  db.Store
	// 用于存储销毁钩子函数，这些函数将关闭插件时调用，比如配置刷新时需要关闭原有的插件实例，其中 key 为随机数
	closeHook map[string]func()
	// closeFunc 函数将在实例销毁时调用
	closeFunc   func() error
	adsHistory  map[string]map[string]struct{}
	pushHisChan chan address
	hisMutex    sync.Mutex
	initFunc    func() error
}

func NewPriDns(config *types.Config, store db.Store) *PriDns {
	closeHook := make(map[string]func())
	adsHistory := make(map[string]map[string]struct{})
	pushHisChan := make(chan address, 1000)
	ticker := time.NewTicker(time.Minute)

	d := &PriDns{
		Config:      config,
		Store:       store,
		closeHook:   closeHook,
		adsHistory:  adsHistory,
		pushHisChan: pushHisChan,
	}

	d.initFunc = func() error {
		// 汇总地址
		go func() {
			for his := range pushHisChan {
				func() {
					d.hisMutex.Lock()
					defer d.hisMutex.Unlock()

					if adsHistory[his.name] == nil {
						adsHistory[his.name] = make(map[string]struct{})
					}
					for _, ip := range his.ads {
						adsHistory[his.name][ip] = struct{}{}
					}
				}()
			}
			fmt.Println("11111111", d)
		}()

		// 定时入库, 每分钟执行一次
		go func() {
			for _ = range ticker.C {
				var mapTmp map[string]map[string]struct{}
				func() {
					d.hisMutex.Lock()
					defer d.hisMutex.Unlock()
					mapTmp, adsHistory = adsHistory, make(map[string]map[string]struct{})
				}()

				// 入库
				ctx := context.Background()
				for name, mp := range mapTmp {
					his := make([]string, len(mp))
					i := 0
					for it, _ := range mp {
						his[i] = it
						i++
					}
					if err := d.Store.SavaHistory(ctx, name, his); err != nil {
						log.Error(err)
					}
				}
			}
			fmt.Println("11111111", d)
		}()
		return nil
	}
	d.closeFunc = func() error {
		for _, f := range closeHook {
			f()
		}
		close(pushHisChan)
		ticker.Stop()
		return nil
	}

	return d
}

// ServeDNS implements the plugin.Handle interface.
func (d PriDns) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	log.Debugf("qname: %s RemoteIp: %s Type: %s QType: %v Class: %s QClass: %v",
		state.Name(), state.IP(), state.Type(), state.QType(), state.Class(), state.QClass())

	// step.1 如果配置了自定义解析，则直接响应配置的自定义解析即可
	answers := handQuery(d, ctx, state)
	if len(answers) != 0 {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true
		m.Answer = answers
		if err := w.WriteMsg(m); err != nil {
			return dns.RcodeServerFailure, err
		}
		return dns.RcodeSuccess, nil
	}

	// step.2 如果没有配置自定义解析则可能需要根据配置将域名转发给特定的DNS服务器进行解析
	if ok, code, err := handForward(d, ctx, state); ok {
		return code, err
	}

	// step.3 如果既没有自定义解析，也没有配置特定的转发，则将请求给下一个插件处理
	return plugin.NextOrFailure(d.Name(), d.Next, ctx, w, r)
}

// Name implements the plugin.Handle interface.
func (d PriDns) Name() string { return "pri-dns" }

// RegisterCloseHook 注册关闭钩子函数，返回一个回调函数用于取消注册
func (d PriDns) RegisterCloseHook(f func()) func() {
	d.hisMutex.Lock()
	defer d.hisMutex.Unlock()

	key := uuid.New().String()
	for {
		if _, ok := d.closeHook[key]; ok {
			key = uuid.New().String()
		} else {
			break
		}
	}
	d.closeHook[key] = f

	return func() {
		d.hisMutex.Lock()
		defer d.hisMutex.Unlock()

		delete(d.closeHook, key)
	}
}

// filterRecord 根据查询域名 qname 过滤掉不需要的数据，并且通过第二个返回值表示是否需要不需要使用全局数据
func filterRecord[T db.RecordFilter](qname string, records []T) ([]T, bool) {
	if len(records) == 0 {
		return nil, false
	}

	names := util.GenAllMatchDomain(qname)
	// 标记是否需要拒绝全局解析
	deny := false
	t := make([]T, 0, len(records))
	for _, name := range names {
		t = t[:0]
		for _, record := range records {
			if record.DenyGlobalVal() {
				deny = true
			} else {
				if record.NameVal() == name {
					t = append(t, record)
				}
			}
		}
		if len(t) != 0 {
			return append(records[:0], t...), false
		}
	}
	return nil, deny
}

//#region query

// handQuery 根据查询名称 qname 和客户端IP remoteIp 查询解析
// 查询规则为：
//  1. 先查询本地添加的解析
//  2. 如果本地没有对应解析则根据规则转发给上游服务器处理
//     如果在规则列表中，则转发到根据规则中指定的上游服务器，否则让下一个插件处理
func handQuery(d PriDns, ctx context.Context, state request.Request) []dns.RR {
	qname := state.Name()
	qname = qname[:len(qname)-1]

	// 目前该插件只处理 IPv4 和 IPv6 查询
	if (state.QType() != dns.TypeA && state.QType() != dns.TypeAAAA) || (state.QClass() != dns.ClassINET) {
		return nil
	}

	// 依次查询私有解析和全局解析
	for _, s := range []string{state.IP(), ""} {
		domains := d.Store.FindDomainByHostAndName(ctx, s, qname)
		// 查询解析,查询后需要过滤掉不需要的（由于查询时一次性查询可能需要的，所以这里需要过滤掉不需要的）
		domains, deny := filterRecord(qname, domains)
		if deny {
			return nil
		}
		if len(domains) == 0 {
			continue
		}

		answers := make([]dns.RR, 0, len(domains))
		for _, domain := range domains {
			switch domain.DnsType {
			case "A":
				r := new(dns.A)
				r.Hdr = dns.RR_Header{Name: qname + ".", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: uint32(domain.Ttl)}
				r.A = net.ParseIP(domain.Value)
				answers = append(answers, r)
			case "AAAA":
				r := new(dns.AAAA)
				r.Hdr = dns.RR_Header{Name: qname + ".", Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: uint32(domain.Ttl)}
				r.AAAA = net.ParseIP(domain.Value)
				answers = append(answers, r)
			default:
				log.Warningf("不支持%s类型!\n", domain.DnsType)
			}
		}

		return answers
	}
	return nil
}

//#endregion

//#region forward

// 尝试处理转发，如果 handForward 已经做出响应（如一个查询需要进行转发或者出现异常情况需要返回），则 ok 为 true，此时直接将 code, err 作为 ServeDNS 返回值即可
// 如果 ok 为 false，则表示 handForward 方法不处理查询，这时一般需要转发给下一个插件处理
func handForward(d PriDns, ctx context.Context, state request.Request) (ok bool, code int, err error) {
	qname := state.Name()
	qname = qname[:len(qname)-1]

	// 依次查询私有解析和全局解析
	for _, s := range []string{state.IP(), ""} {
		forwards := d.Store.FindForwardByHostAndName(ctx, s, qname)
		// 查询转发,查询后需要过滤掉不需要的（由于查询时一次性查询可能需要的，所以这里需要过滤掉不需要的）
		forwards, deny := filterRecord(qname, forwards)
		if deny {
			return
		}
		if len(forwards) == 0 {
			continue
		}
		ok = true

		// 查询对应的 Proxy 实例
		proxies, err2 := myForward.GetProxy(d.Config, forwards, d.RegisterCloseHook)
		if err2 != nil {
			code = dns.RcodeServerFailure
			err = err2
			return
		}

		// 转发请求
		code2, err2, rrs := myForward.Run(proxies, ctx, state)
		code = code2
		err = err2

		// 存储解析历史
		if rrs != nil {
			d.pushHisChan <- address{name: forwards[0].Name, ads: rrs}
		}
		return
	}

	return
}

//#endregion

// 规划化域名 '.' 'example.com.' - _ = plugin.Host("example.com.").NormalizeExact()[0]
