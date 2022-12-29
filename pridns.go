package pridns

import (
	"context"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// Config 表示插件配置
type Config struct {
	// DataSourceName 为MySQL数据库连接地址
	DataSourceName string
}

type PriDns struct {
	Config *Config
	Next   plugin.Handler
}

// ServeDNS implements the plugin.Handle interface.
func (d PriDns) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	qname := state.Name()
	_ = qname

	var answers []dns.RR

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
func (d PriDns) Name() string { return "pridns" }
