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
			hosts = mergeIpByMultiMaskAndLevel(hosts, arr, ctx.URLParamBoolDefault("pri", true))

			ctx.WriteString(strings.Join(hosts, ","))
		}
		apiParty.Get("/ip-line", getIpLine)
		apiParty.Get("/ip-line.txt", getIpLine)
	}

	return app
}

// 私有地址
var priNets = []net.IPNet{
	{IP: net.ParseIP("10.0.0.0"), Mask: net.CIDRMask(8, 8*net.IPv4len)},
	{IP: net.ParseIP("172.16.0.0"), Mask: net.CIDRMask(12, 8*net.IPv4len)},
	{IP: net.ParseIP("192.168.0.0"), Mask: net.CIDRMask(16, 8*net.IPv4len)},
}

func mergeIpByMultiMaskAndLevel(hosts []string, arr [][2]int, pri bool) []string {
	// 解析IP地址，并去除私有地址
	hostIPNets := make([]*net.IPNet, len(hosts))
	i := 0
HostFor:
	for _, host := range hosts {
		// 如果不是 CIDR 格式则在末尾拼接掩码将其转换为 CIDR 格式
		if strings.IndexByte(host, '/') == -1 {
			host = fmt.Sprintf("%s/%d", host, 32)
		}

		_, ipNet, err := net.ParseCIDR(host)
		if err != nil {
			continue
		}
		// 可能需要排除私有地址
		if pri {
			for _, priNet := range priNets {
				if priNet.Contains(ipNet.IP) {
					continue HostFor
				}
			}
		}
		hostIPNets[i] = ipNet
		i++
	}
	hostIPNets = hostIPNets[:i]

	// 根据规则进行合并
	dst := make([]*net.IPNet, 0, len(hostIPNets)*len(arr))
	for _, it := range arr {
		dst = append(dst, mergeIpByMaskAndLevel(hostIPNets, it[0], it[1])...)
	}

	// 排序和去重
	ipNets := cidr_merger.MergeIPNet(dst, false)
	ips := make([]string, len(ipNets))
	for i, ipNet := range ipNets {
		ips[i] = ipNet.String()
	}
	return ips
}

func mergeIpByMaskAndLevel(hosts []*net.IPNet, mask, level int) []*net.IPNet {
	l := len(hosts)

	ipNets := make([]*net.IPNet, l) // 目标网络地址
	j := 0
	for _, host := range hosts {
		size, _ := host.Mask.Size()
		if mask < size {
			mask := net.CIDRMask(mask, 8*net.IPv4len)
			ipNets[j] = &net.IPNet{
				IP:   host.IP.Mask(mask),
				Mask: mask,
			}
		} else {
			ipNets[j] = host
		}
		j++
	}
	ipNets = ipNets[:j]
	ipNets = cidr_merger.MergeIPNet(ipNets, false)

	// 将原始ip根据目标网络地址进行分组
	ipMap := make(map[string][]*net.IPNet, len(ipNets))
	for _, item := range hosts {
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
	dst := make([]*net.IPNet, 0, j)
	for netKey, nets := range ipMap {
		if nets != nil {
			if len(nets) >= level {
				_, ipNet, _ := net.ParseCIDR(netKey)
				dst = append(dst, ipNet)
			} else {
				for _, it := range nets {
					dst = append(dst, it)
				}
			}
		}
	}

	return dst
}
