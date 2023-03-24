package pri_dns

import (
	"context"
	"fmt"
	"github.com/kataras/iris/v12"
	cidrMerger "github.com/laeni/pri-dns/cidr-merger"
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
			pri := ctx.URLParamBoolDefault("pri", true)
			v := ctx.URLParamIntDefault("v", 2)

			hosts := store.FindHistoryByHost(context.Background(), addr)
			{
				// 解析IP地址，并去除私有地址
				hostIPNets := strToIpNet(hosts, pri)

				// 根据版本进行合并
				switch v {
				case 1: // mask=8&level=100 & mask=12&level=50 & mask=16&level=10 & mask=24&level=1
					query := ctx.Request().URL.Query()
					masks, maskOk := query["mask"]    // 掩码位数. 取值为 1-32
					levels, levelOk := query["level"] // 在一个特定的网段下（网段由具体某个IP和mask决定），如果存在的解析历史数量达到该值，则直接取该网段的网络地址
					var arr [][2]int
					if maskOk || levelOk {
						if len(masks) == len(levels) { // 解析
							for i := 0; i < len(masks); i++ {
								mask, maskErr := strconv.Atoi(masks[i])
								level, levelErr := strconv.Atoi(levels[i])
								if maskErr != nil || levelErr != nil || mask < 1 || mask > 32 || level < 1 {
									ctx.StopWithStatus(http.StatusBadRequest)
									return
								} else {
									arr = append(arr, [2]int{mask, level})
								}
							}
						} else { // 异常
							ctx.StopWithStatus(http.StatusBadRequest)
							return
						}
					} else {
						//                                  256       128       64        32       16       8        4
						arr = [][2]int{{8, 100}, {16, 50}, {24, 25}, {25, 20}, {26, 10}, {27, 5}, {28, 4}, {29, 3}, {30, 2}}
					}

					hostIPNets = mergeIpV1(hostIPNets, pri, arr)
				case 2: // level24=1 & level16=5 & level8=10
					query := ctx.Request().URL.Query()
					levels, levelOk := query["level"]
					var level [3]int
					if levelOk {
						if len(levels) == 3 { // 解析
							for i := 0; i < 3; i++ {
								levelInt, levelErr := strconv.Atoi(levels[i])
								if levelErr != nil || levelInt < 1 {
									ctx.StopWithStatus(http.StatusBadRequest)
									return
								} else {
									level[i] = levelInt
								}
							}
						} else { // 异常
							ctx.StopWithStatus(http.StatusBadRequest)
							return
						}
					} else {
						level = [...]int{1, 3, 10}
					}

					hostIPNets = mergeIpV2(hostIPNets, pri, level)
				default:
					ctx.StopWithStatus(http.StatusBadRequest)
					return
				}

				// 排序、去重、转为String
				hosts = ipNetToString(cidrMerger.MergeIPNet(hostIPNets, false))
			}

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

// mergeIpV1 V1版本的合并规则，明确通过指定将达到指定数量的原始IP进行合并，
// 比如 'mask=16&level=10' 表示如果存在10个ip在16位掩码时的网络地址相同，则用16位掩码的网络地址表示它们。
// mask和level可以出现多次，且必须成对出现，当出现多次时，分别用他们和原始IP进行计算，并将每次得到的结果在最后进行合并。
// pri 参数表示默认情况下是否需要过滤掉内网IP
func mergeIpV1(hostIPNets []*net.IPNet, pri bool, arr [][2]int) []*net.IPNet {
	dst := make([]*net.IPNet, 0, len(hostIPNets)*len(arr))
	for _, it := range arr {
		dst = append(dst, mergeIpByMaskAndLevel(hostIPNets, it[0], it[1], pri)...)
	}
	return dst
}

// mergeIpV2 对IP进行简单合并，合并分为3步：
// 第1步，对原始数据进行合并，合并规则和版本1一样，但此步骤中明确指定mask为24，level值为第1个level参数的值
// 第2步，合并规则和第1步一样，但此步骤的输入数据不再是原始IP，而是第1步中生成的结果，且明确指定mask为16，level值为第2个level参数的值
// 第3步，再次重复第2步，此步骤的输入数据也不是原始IP，而是第2步中生成的结果，且明确指定mask为8，level值为第3个level参数的值。此步骤得到的结果为最终结果
func mergeIpV2(hostIPNets []*net.IPNet, pri bool, level [3]int) []*net.IPNet {
	// 从第2步开始，需要把不符合进入下一步的数据筛选出来
	other := make([]*net.IPNet, 0, len(hostIPNets))

	// 第1步
	hostIPNets = mergeIpByMaskAndLevel(hostIPNets, 24, level[0], pri)

	// 第2步
	i := 0
	for _, hostIPNet := range hostIPNets {
		maskSize, _ := hostIPNet.Mask.Size()
		if maskSize > 24 {
			other = append(other, hostIPNet)
		} else {
			hostIPNets[i] = hostIPNet
			i++
		}
	}
	hostIPNets = mergeIpByMaskAndLevel(hostIPNets[:i], 16, level[1], pri)

	// 第3步
	i = 0
	for _, hostIPNet := range hostIPNets {
		maskSize, _ := hostIPNet.Mask.Size()
		if maskSize > 16 {
			other = append(other, hostIPNet)
		} else {
			hostIPNets[i] = hostIPNet
			i++
		}
	}
	hostIPNets = mergeIpByMaskAndLevel(hostIPNets[:i], 8, level[1], pri)

	return append(hostIPNets, other...)
}

// 将String格式的IP转换为网段对象
func strToIpNet(hosts []string, pri bool) []*net.IPNet {
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
	return hostIPNets[:i]
}

// 将网段对象格式的IP转换为String
func ipNetToString(ipNets []*net.IPNet) []string {
	ips := make([]string, len(ipNets))
	for i, ipNet := range ipNets {
		ips[i] = ipNet.String()
	}
	return ips
}

func mergeIpByMaskAndLevel(hosts []*net.IPNet, mask, level int, pri bool) []*net.IPNet {
	l := len(hosts)

	ipNets := make([]*net.IPNet, l) // 目标网络地址
	j := 0
HostFor:
	for _, host := range hosts {
		size, _ := host.Mask.Size()
		if mask < size {
			mask := net.CIDRMask(mask, 8*net.IPv4len)
			tarIpNet := &net.IPNet{IP: host.IP.Mask(mask), Mask: mask}
			// 如果需要过滤内网IP，则当指定掩码的时候，生成的IP不得是内网IP，如果是，则需要选择一个更长的掩码以满足要求
			if pri {
				priMask, ok := isPriIpNet(tarIpNet)
				for ok {
					priMaskSize, _ := priMask.Size()
					if priMaskSize < size {
						tarIpNet = &net.IPNet{IP: host.IP.Mask(priMask), Mask: priMask}
						// 新生成的网段有可能还是包含内网地址，所以还需要继续处理（由于前面已经过滤掉内网地址，所以这里当掩码长度大于等于内网掩码时一定存在不不包含内网的网段）
						priMask, ok = isPriIpNet(tarIpNet)
					} else {
						ipNets[j] = host
						j++
						continue HostFor
					}
				}
			}
			ipNets[j] = tarIpNet
			j++
		} else {
			ipNets[j] = host
			j++
		}
	}
	ipNets = ipNets[:j]
	ipNets = cidrMerger.MergeIPNet(ipNets, false)

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

// isPriIpNet 判断目标网络是否包含内网地址，如果是的话，返回包含的内网的掩码
func isPriIpNet(ipNet *net.IPNet) (net.IPMask, bool) {
	for _, priIpNet := range priNets {
		if ipNet.Contains(priIpNet.IP) {
			return priIpNet.Mask, true
		}
	}
	return nil, false
}
