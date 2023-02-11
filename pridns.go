package pridns

import (
	"context"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/laeni/pri-dns/db"
	"github.com/laeni/pri-dns/util"
	"github.com/miekg/dns"
	"net"
)

var log = clog.NewWithPlugin("pri-dns")

const (
	storeTypeMySQL = "mysql" // 表示使用 mysql 作为存储介质
	storeTypeEtcd  = "etcd"  // 表示使用 etcd 作为存储介质
	storeTypeRedis = "redis" // 表示使用 redis 作为存储介质

	dnsTypeDENY = "DENY" // 自定义类型，表示用于拒绝全局解析
)

type MySQLConfig struct {
	// DataSourceName 为MySQL数据库连接地址
	dataSourceName string
}

// Config 表示插件配置
type Config struct {
	adminPassword string
	storeType     string
	mySQL         MySQLConfig
}

type PriDns struct {
	Config *Config
	Next   plugin.Handler
	Store  db.Store
}

// ServeDNS implements the plugin.Handle interface.
func (d PriDns) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := state.Name()

	log.Infof("qname: %s RemoteIp: %s Type: %s QType: %v Class: %s QClass: %v",
		qname, state.IP(), state.Type(), state.QType(), state.Class(), state.QClass())

	answers := handQuery(d, state, qname[:len(qname)-1], state.IP())

	if len(answers) == 0 {
		return plugin.NextOrFailure(d.Name(), d.Next, ctx, w, r)
	}

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	m.Answer = answers
	if err := w.WriteMsg(m); err != nil {
		return dns.RcodeServerFailure, err
	}

	return dns.RcodeSuccess, nil
}

// Name implements the plugin.Handle interface.
func (d PriDns) Name() string { return "pri-dns" }

//#region query

// handQuery 根据查询名称 qname 和客户端IP remoteIp 查询解析
// 查询规则为：
//  1. 先查询本地添加的解析
//  2. 如果本地没有对应解析则根据规则转发给上游服务器处理
//     如果在规则列表中，则转发到根据规则中指定的上游服务器，否则让下一个插件处理
func handQuery(d PriDns, state request.Request, qname string, remoteIp string) []dns.RR {
	// 目前该插件只处理 IPv4 和 IPv6 查询
	if (state.QType() != dns.TypeA && state.QType() != dns.TypeAAAA) || (state.QClass() != dns.ClassINET) {
		return nil
	}

	// 依次查询私有解析和全局解析
	for _, s := range []string{remoteIp, ""} {
		domains := d.Store.FindDomainByHostAndName(s, qname)
		// 查询私有解析,查询后需要过滤掉不需要的（由于查询时一次性查询可能需要的，所以这里需要过滤掉不需要的）
		domains, deny := filterDomain(qname, domains)
		if len(domains) == 0 || deny {
			return nil
		}

		answers := make([]dns.RR, 0, len(domains))
		for _, domain := range domains {
			switch domain.Type {
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
				log.Warningf("不支持%s类型!\n", domain.Type)
			}
		}

		return answers
	}
	return nil
}

// filterDomain 根据查询域名 qname 过滤掉不需要的解析
func filterDomain(qname string, d []db.Domain) ([]db.Domain, bool) {
	if len(d) == 0 {
		return nil, false
	}

	names := util.GenAllMatchDomain(qname)
	// 标记是否需要拒绝全局解析
	deny := false
	t := make([]db.Domain, 0, len(d))
	for _, name := range names {
		t = t[:0]
		for _, domain := range d {
			if domain.Type != dnsTypeDENY {
				if domain.Name == name {
					t = append(t, domain)
				}
			} else {
				deny = true
			}
		}
		if len(t) != 0 {
			return append(d[:0], t...), false
		}
	}
	return nil, deny
}

//#endregion

//#region forward
//#endregion
