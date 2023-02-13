// Package forward implements a forwarding proxy. It caches an upstream net.Conn for some time, so if the same
// client returns the upstream's Conn will be precached. Depending on how you benchmark this looks to be
// 50% faster than just opening a new connection for every client. It works with UDP and TCP and uses
// inband healthchecking.
package forward

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/coredns/coredns/plugin/pkg/parse"
	"github.com/laeni/pri-dns/db"
	"github.com/laeni/pri-dns/util"
	"sort"
	"strings"
	"time"

	"github.com/coredns/coredns/plugin/debug"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"

	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin("pri-dns")

var (
	// max_fails 是考虑之前需要的后续失败运行状况检查的数量 上游要下降。如果为 0，则上游永远不会标记为关闭（也不进行健康检查）。 默认值为 2。
	maxFails uint32 = 2
	// 转发处理超时时间
	defaultTimeout = 5 * time.Second

	// 健康检查间隔
	hcInterval    = 500 * time.Millisecond
	maxDnsSvr     = 4  // 单条记录转发上游的最大数量。
	maxProxyCache = 20 // 最大代理实例缓存数量

	// 访问远程 TLS 上游的配置
	tlsConfig = &tls.Config{
		// 默认情况下，ClientSessionCache 为 nil，即不会从历史会话中恢复，所以如果有 ClientSessionCache，则二次连接同一 TLS 服务器时握手可能会加快。
		ClientSessionCache: tls.NewLRUClientSessionCache(0),
	}

	// ErrNoHealthy means no healthy proxies left.
	ErrNoHealthy = errors.New("no healthy proxies")
	// ErrNoForward means no forwarder defined.
	ErrNoForward = errors.New("no forwarder defined")
	// ErrCachedClosed means cached connection was closed by peer.
	ErrCachedClosed = errors.New("cached connection was closed by peer")
)

// Run 使用代理实际进行转发.
// 该代码几乎复制于 forward 插件
func Run(proxies []*Proxy, ctx context.Context, state request.Request) (int, error) {
	fails := 0
	var upstreamErr error
	i := 0
	list := util.RandomList(proxies)
	deadline := time.Now().Add(defaultTimeout)
	for time.Now().Before(deadline) {
		// 如果没成功且还有时间则重试
		if i >= len(list) {
			// reached the end of list, reset to begin
			i = 0
			fails = 0
		}

		proxy := list[i]
		i++
		// 跳过健康检查未通过的（最后一次不跳过，因为即使健康检查即使不通过也可能可以正常使用）
		if proxy.Down(maxFails) {
			fails++
			if fails < len(list) {
				continue
			}
			HealthcheckBrokenCount.Add(1)
		}

		var (
			ret *dns.Msg
			err error
		)
		for {
			ret, err = proxy.Connect(ctx, state, false, false)

			if err == ErrCachedClosed { // Remote side closed conn, can only happen with TCP.
				continue
			}
			break
		}

		upstreamErr = err

		// 如果上游错误则进行一次健康检测，并且如果还有上游的话就继续使用其他上游，否则结束
		if err != nil {
			// Kick off health check to see if *our* upstream is broken.
			if maxFails != 0 {
				proxy.Healthcheck()
			}

			if fails < len(list) {
				continue
			}
			break
		}

		// Check if the reply is correct; if not return FormErr.
		// 检查上游回复的响应是否正确，如果不正确则返回 dns.RcodeFormatError 错误
		if !state.Match(ret) {
			debug.Hexdumpf(ret, "Wrong reply for id: %d, %s %d", ret.Id, state.QName(), state.QType())

			formerr := new(dns.Msg)
			formerr.SetRcode(state.Req, dns.RcodeFormatError)
			state.W.WriteMsg(formerr)
			return dns.RcodeSuccess, nil
		}

		_ = state.W.WriteMsg(ret)
		return dns.RcodeSuccess, nil
	}

	if upstreamErr != nil {
		return dns.RcodeServerFailure, upstreamErr
	}

	return dns.RcodeServerFailure, ErrNoHealthy
}

//#region ProxyCache

type proxyCacheWrapper struct {
	dnsSvr       string    // 上游DNS地址
	activityTime time.Time // 最后活跃时间，如果不活跃的可能会从缓存中剔除
	proxy        *Proxy
	closeHook    func()
}

type proxyCacheType []*proxyCacheWrapper

func (pc proxyCacheType) get(name string) *proxyCacheWrapper {
	for _, it := range pc {
		if it.dnsSvr == name {
			return it
		}
	}
	return nil
}

func (pc proxyCacheType) Len() int {
	return len(pc)
}

func (pc proxyCacheType) Less(i, j int) bool {
	return pc[i].activityTime.Before(pc[j].activityTime)
}

func (pc proxyCacheType) Swap(i, j int) {
	pc[i], pc[j] = pc[j], pc[i]
}

var proxyCache = make(proxyCacheType, 0, maxProxyCache)

//#endregion

// GetProxy 根据 forwards 数据从缓存查询对应的 Proxy 实例，如果缓存不存在则创建新实例加入缓存
func GetProxy(forwards []db.Forward, RegisterCloseHook func(func()) func()) []*Proxy {
	var proxies []*Proxy
	for _, proxy := range forwards {
		// 规范化DNS地址
		toHosts, err := parse.HostPortOrFile(strings.Join(proxy.DnsSvr, " "))
		if err != nil {
			continue
		}

		for _, svr := range toHosts {
			if len(proxies) >= maxDnsSvr {
				return proxies
			}

			wrapper := proxyCache.get(svr)
			// 如果不存在则创建新的代理实例放入缓存
			if wrapper == nil {
				p := newProxy(svr)
				wrapper = &proxyCacheWrapper{
					dnsSvr:       svr,
					activityTime: time.Now(),
					proxy:        p,
					closeHook:    RegisterCloseHook(p.stop),
				}
				// 如果缓存已经超过最大限制，则删除不活跃的
				if len(proxyCache) > maxProxyCache {
					sort.Sort(proxyCache)
					proxyCacheTmp, removed := proxyCache[:len(proxyCache)-5], proxyCache[len(proxyCache)-5:]
					proxyCache = proxyCacheTmp
					// 移除的代理需要关闭健康检查
					for _, it := range removed {
						it.proxy.stop()
						// 已经关闭的Proxy要取消注册
						it.closeHook()
					}
				}
				proxyCache = append(proxyCache, wrapper)
			} else {
				wrapper.activityTime = time.Now()
			}
			proxies = append(proxies, wrapper.proxy)
		}
	}
	return proxies
}

func newProxy(dnsSvr string) *Proxy {
	trans, h := parse.Transport(dnsSvr)
	p := NewProxy(h, trans)
	if trans == "tls" {
		p.SetTLSConfig(tlsConfig)
	}
	// 在此时间后过期（缓存）连接
	p.SetExpire(10 * time.Second)

	// 启动健康检查
	p.start(hcInterval)

	return p
}
