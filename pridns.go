package pri_dns

import (
	"context"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/google/uuid"
	"github.com/laeni/pri-dns/db"
	myForward "github.com/laeni/pri-dns/forward"
	"github.com/laeni/pri-dns/types"
	"github.com/miekg/dns"
	"net"
	"strings"
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
	pushHisChan chan address
	hisMutex    sync.Mutex
	initFunc    func() error
}

func NewPriDns(config *types.Config, store db.Store) *PriDns {
	closeHook := make(map[string]func())
	// 域名和IP. {"laeni.cn": {"127.0.0.1":nil, "127.0.0.2":nil}}
	adsHistory := make(map[string]map[string]struct{})
	pushHisChan := make(chan address, 1000)
	ticker := time.NewTicker(time.Minute)

	d := &PriDns{
		Config:      config,
		Store:       store,
		closeHook:   closeHook,
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
		}()

		// 定时入库, 每分钟执行一次
		go func() {
			for range ticker.C {
				var mapTmp map[string]map[string]struct{}
				func() {
					d.hisMutex.Lock()
					defer d.hisMutex.Unlock()
					mapTmp, adsHistory = adsHistory, make(map[string]map[string]struct{})
				}()

				// 入库
				for name, mp := range mapTmp {
					his := make([]string, len(mp))
					i := 0
					for it := range mp {
						his[i] = it
						i++
					}
					if err := d.Store.SavaHistory(name, his); err != nil {
						log.Error(err)
					}
				}
			}
		}()
		return nil
	}
	d.closeFunc = func() error {
		for key, f := range closeHook {
			f()
			delete(closeHook, key)
		}
		close(pushHisChan)
		ticker.Stop()
		return nil
	}

	return d
}

// ServeDNS implements the plugin.Handle interface.
func (d *PriDns) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	log.Debugf("qname: %s RemoteIp: %s Type: %s QType: %v Class: %s QClass: %v",
		state.Name(), state.IP(), state.Type(), state.QType(), state.Class(), state.QClass())

	// step.1 如果配置了自定义解析，则直接响应配置的自定义解析即可
	answers := handQuery(d, state)
	if len(answers) != 0 {
		log.Debugf("已找到自定义解析记录: %v", answers)
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
func (d *PriDns) Name() string { return "pri-dns" }

// RegisterCloseHook 注册关闭钩子函数，返回一个回调函数用于取消注册; 如果钩子函数被调用，则也会自动取消注册
func (d *PriDns) RegisterCloseHook(f func()) func() {
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

// filterRecord 根据查询域名 qname 及优先级找一个最佳的
func filterRecord(records []db.Forward) db.Forward {
	if len(records) == 0 {
		panic("需要过滤的数据不能为空")
	}

	t := records[0]
	for i := 1; i < len(records); i++ {
		if matchPriorityCompare(records[i], t) > 0 {
			t = records[i]
		}
	}

	return t
}

// filterDomain 根据查询域名 qname 及优先级找最佳的解析，同一个域名的解析记录可能由多个
func filterDomain(domains []db.Domain) map[string][]db.Domain {
	if len(domains) == 0 {
		return nil
	}

	// 根据解析类型分类（A、AAAA等）
	domainByDnsType := make(map[string][]db.Domain)
	for _, domain := range domains {
		if domainByDnsType[domain.DnsType] == nil {
			domainByDnsType[domain.DnsType] = []db.Domain{domain}
		} else {
			domainByDnsType[domain.DnsType] = append(domainByDnsType[domain.DnsType], domain)
		}
	}

	// 每个分类中，根据优先级选择最匹配的解析记录，优先级为：私有解析 > 全局解析; 精准匹配 > 泛解析; 精度高的泛解析 > 精度低的泛解析
	for dnsType, items := range domainByDnsType {
		if len(items) <= 1 {
			continue
		}
		var slice []db.Domain
		for _, item := range items {
			if len(slice) == 0 {
				slice = []db.Domain{item}
			} else {
				compare := matchPriorityCompare(item, slice[0])
				if compare > 0 {
					slice = []db.Domain{item}
				}
				if compare == 0 {
					slice = append(slice, item)
				}
				// compare < 0 时直接丢弃
			}
		}
		domainByDnsType[dnsType] = slice
	}
	// 如果由拒绝策略，则忽略
	for dnsType, items := range domainByDnsType {
		for _, item := range items {
			if item.DenyGlobal {
				delete(domainByDnsType, dnsType)
			}
		}
	}

	return domainByDnsType
}

// 根据域名匹配规则比较 a 和 b 的优先级，优先级为：私有解析 > 全局解析; 精准匹配 > 泛解析; 精度高的泛解析 > 精度低的泛解析，
// 如果 a 优先级高于 b，则返回 1；如果 a 和 b 优先级相同则返回 0；如果 a 优先级低于 b 则返回 -1
func matchPriorityCompare(a, b db.RecordFilter) int {
	// 私有 > 全局
	if a.ClientHostVal() != "" && b.ClientHostVal() == "" {
		return 1
	}
	if a.ClientHostVal() == "" && b.ClientHostVal() != "" {
		return -1
	}
	aName := a.NameVal()
	bName := b.NameVal()
	// 精准解析 > 泛解析
	if !strings.Contains(aName, "*") && strings.Contains(bName, "*") {
		return 1
	}
	if strings.Contains(aName, "*") && !strings.Contains(bName, "*") {
		return -1
	}
	// 高精度 > 低精度
	if len(aName) > len(bName) {
		return 1
	}
	if len(aName) < len(bName) {
		return -1
	}
	return 0
}

// region query

// handQuery 根据查询名称 qname 和客户端IP remoteIp 查询解析
// 查询规则为：
//  1. 先查询本地添加的解析
//  2. 如果本地没有对应解析则根据规则转发给上游服务器处理
//     如果在规则列表中，则转发到根据规则中指定的上游服务器，否则让下一个插件处理
func handQuery(d *PriDns, state request.Request) []dns.RR {
	qname := state.Name()
	qname = qname[:len(qname)-1]

	// 目前该插件只处理 IPv4 和 IPv6 查询
	if (state.QType() != dns.TypeA && state.QType() != dns.TypeAAAA) || (state.QClass() != dns.ClassINET) {
		return nil
	}

	// 一次查询私有解析（clientHost 对应的数据）和全局解析（clientHost 对空的数据）
	domains := d.Store.FindDomainByHostAndName(state.IP(), qname)
	if len(domains) == 0 {
		return nil
	}
	// 根据优先级找到最匹配的一个
	domainByType := filterDomain(domains)
	if len(domainByType) == 0 {
		return nil
	}

	answers := make([]dns.RR, 0, len(domains))
	for tp, items := range domainByType {
		switch tp {
		case "A":
			for _, domain := range items {
				r := new(dns.A)
				r.Hdr = dns.RR_Header{Name: qname + ".", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: uint32(domain.Ttl)}
				r.A = net.ParseIP(domain.Value)
				answers = append(answers, r)
			}
		case "AAAA":
			for _, domain := range items {
				r := new(dns.AAAA)
				r.Hdr = dns.RR_Header{Name: qname + ".", Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: uint32(domain.Ttl)}
				r.AAAA = net.ParseIP(domain.Value)
				answers = append(answers, r)
			}
		default:
			log.Warningf("不支持%s类型!\n", tp)
		}
	}

	return answers
}

// endregion

// region forward

// 尝试处理转发，如果 handForward 已经做出响应（如一个查询需要进行转发或者出现异常情况需要返回），则 ok 为 true，此时直接将 code, err 作为 ServeDNS 返回值即可
// 如果 ok 为 false，则表示 handForward 方法不处理查询，这时一般需要转发给下一个插件处理
func handForward(d *PriDns, ctx context.Context, state request.Request) (ok bool, code int, err error) {
	qname := state.Name()
	qname = qname[:len(qname)-1]

	// 一次查询私有转发（clientHost 对应的数据）和全局转发（clientHost 对空的数据）
	forwards := d.Store.FindForwardByHostAndName(state.IP(), qname)
	if len(forwards) == 0 {
		return
	}
	// 根据优先级找到最合适的个转发配置
	forward := filterRecord(forwards)
	if forward.DenyGlobal {
		log.Debug("没有有效的转发记录")
		return
	}
	log.Debugf("解析转发: %s => %v", qname, forward.DnsSvr)
	ok = true

	// 查询对应的 Proxy 实例
	proxies, err2 := myForward.GetProxy(d.Config, forward.DnsSvr, d.RegisterCloseHook)
	if err2 != nil {
		code = dns.RcodeServerFailure
		err = err2
		return
	}

	// 转发请求
	var rrs []string
	code, err, rrs = myForward.Run(proxies, ctx, state)

	if rrs != nil {
		log.Debugf("解析结果: %v", rrs)
		// 存储解析历史
		d.pushHisChan <- address{name: forwards[0].Name, ads: rrs}
	}
	return
}

// endregion

// 规划化域名 '.' 'example.com.' - _ = plugin.Host("example.com.").NormalizeExact()[0]
