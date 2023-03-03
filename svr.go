package pridns

import (
	"context"
	"fmt"
	"github.com/kataras/iris/v12"
	cidr_merger "github.com/laeni/pri-dns/cidr-merger"
	"github.com/laeni/pri-dns/db"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	app *iris.Application // 应用程序实例
)

// StartApp 启动服务。启动时需要 PriDns 实例来注册关闭回调钩子以及读取相关配置
func StartApp(p *PriDns) error {
	// 目前不支持修改应用配置
	if app != nil {
		return nil
	}
	app = newApp(p.Store)

	var startError error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		startError = func() error {
			var err error

			// 检测服务是否正常启动
			go func() {
				var healthErr error
				for i := 0; i < 100; i++ {
					if err != nil {
						wg.Done()
						return
					}
					resp, errRes := http.Get(fmt.Sprintf("http://127.0.0.1:%s/health", strings.Split(p.Config.ServerPort, ":")[1]))
					if errRes == nil && (resp.StatusCode >= 200 && resp.StatusCode < 300) {
						wg.Done()
						return
					}
					healthErr = errRes
					time.Sleep(100 * time.Millisecond)
				}
				log.Error("后台服务健康检查无法通过", healthErr)
			}()

			err = app.Listen(p.Config.ServerPort)
			return err
		}()
	}()
	wg.Wait()
	return startError
}

func newApp(store db.Store) *iris.Application {
	app = iris.New()
	app.Get("/health", func(c iris.Context) {
		_, _ = c.WriteString("OK")
	})

	apiParty := app.Party("/api")
	{
		getIpLine := func(ctx iris.Context) {
			addr := ctx.RemoteAddr()

			// mask=8&level=100 mask=8&level=100
			query := ctx.Request().URL.Query()
			masks, maskOk := query["mask"]    // 掩码位数. 取值为 1-32
			levels, levelOk := query["level"] // 在一个特定的网段下（网段由具体某个IP和mask决定），如果存在的解析历史数量达到该值，则直接取该网段的网络地址
			var arr [][2]int
			if maskOk || levelOk {
				if len(masks) == len(levels) { // 解析
					for i := 0; i < len(masks); i++ {
						mask, maskErr := strconv.Atoi(masks[i])
						level, levelErr := strconv.Atoi(levels[i])
						if maskErr != nil || levelErr != nil || mask < 1 || mask > 24 || level < 1 {
							ctx.StopWithStatus(http.StatusBadRequest)
						} else {
							arr = append(arr, [2]int{mask, level})
						}
					}
				} else { // 异常
					ctx.StopWithStatus(http.StatusBadRequest)
				}
			} else {
				//                                  256       128       64        32       16       8        4
				arr = [][2]int{{8, 100}, {16, 50}, {24, 25}, {25, 20}, {26, 10}, {27, 5}, {28, 4}, {29, 3}, {30, 2}}
			}

			hosts := store.FindHistoryByHost(context.Background(), addr)
			pri := ctx.URLParamBoolDefault("pri", true)
			hosts = mergeIpByMultiMaskAndLevel(hosts, arr, pri)

			ctx.WriteString(strings.Join(hosts, ","))
		}
		apiParty.Get("/ip-line", getIpLine)
		apiParty.Get("/ip-line.txt", getIpLine)
	}

	return app
}

type IPNetWrapper struct {
	IPNet  *net.IPNet
	String string
}

// 排除私有地址
var priNets = []net.IPNet{
	{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 8*net.IPv4len)},
	{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 8*net.IPv4len)},
	{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 8*net.IPv4len)},
}

func mergeIpByMultiMaskAndLevel(hosts []string, arr [][2]int, pri bool) []string {
	dst := make([]string, 0, len(hosts)*len(arr))

	for _, it := range arr {
		dst = append(dst, mergeIpByMaskAndLevel(hosts, it[0], it[1], pri)...)
	}

	return cidr_merger.MergeIp(dst, false)
}

func mergeIpByMaskAndLevel(hosts []string, mask, level int, pri bool) []string {
	l := len(hosts)

	hostWrappers := make([]*net.IPNet, l) // 原始IP
	ipNets := make([]*net.IPNet, l)       // 目标网络地址
	j := 0
ForHosts:
	for _, host := range hosts {
		ip, ipNet, err := net.ParseCIDR(host)
		if err != nil {
			continue
		}
		// 可能需要排除私有地址
		if pri {
			for _, priNet := range priNets {
				if priNet.Contains(ip) {
					continue ForHosts
				}
			}
		}

		hostTmp := host

		// 预处理 - 1.如果不是网络地址的，将其转换为网络地址 2.如果目标掩码位数小于实际掩码位数的，以目标掩码位数为准
		idx := strings.IndexByte(hostTmp, '/')
		if idx == -1 {
			hostTmp = fmt.Sprintf("%s/%d", hostTmp, mask)
		} else {
			addr, maskStr := hostTmp[:idx], hostTmp[idx+1:]
			mask2, err := strconv.Atoi(maskStr)
			if err != nil {
				continue
			}
			if mask < mask2 {
				hostTmp = fmt.Sprintf("%s/%d", addr, mask)
			}
		}

		_, ipNetTmp, err := net.ParseCIDR(hostTmp)
		if err != nil {
			continue
		}

		hostWrappers[j] = ipNet
		ipNets[j] = ipNetTmp
		j++
	}
	hostWrappers = hostWrappers[:j]
	ipNets = ipNets[:j]
	ipNets = cidr_merger.MergeIPNet(ipNets, false)

	// 将原始ip根据目标网络地址进行分组
	ipMap := make(map[string][]*net.IPNet, len(ipNets))
	for _, item := range hostWrappers {
		for _, ipNet := range ipNets {
			if ipNet.Contains(item.IP) {
				key := ipNet.String()
				_, ok := ipMap[key]
				if !ok {
					ipMap[key] = []*net.IPNet{item}
				} else {
					ipMap[key] = append(ipMap[key], item)
				}
			}
		}
	}

	// 根据 level 指定的数量觉得是否取原始数据还是他们对应期望网络地址
	dst := make([]string, 0, j)
	for netKey, nets := range ipMap {
		if nets != nil {
			if len(nets) >= level {
				dst = append(dst, netKey)
			} else {
				for _, it := range nets {
					dst = append(dst, it.String())
				}
			}
		}
	}

	return dst
}
