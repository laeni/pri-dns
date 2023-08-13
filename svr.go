package pri_dns

import (
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
		// 获取 WireGuard 代理IP
		getIpLine := func(ctx iris.Context) {
			addr := ctx.RemoteAddr()
			v := ctx.URLParamIntDefault("v", 2)

			hosts, hisExs := store.FindHistoryByHost(addr)
			{
				// 解析IP地址
				hostIPNets := cidrMerger.StrToIpNet(hosts)

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

					hostIPNets = mergeIpV1(hostIPNets, arr)
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

					hostIPNets = mergeIpV2(hostIPNets, level)
				default:
					ctx.StopWithStatus(http.StatusBadRequest)
					return
				}

				ipRange := cidrMerger.IpNetToRange(hostIPNets)
				// 排序、去重（主要目的是减少数量）
				ipRange = cidrMerger.SortAndMerge(ipRange)
				// 排除指定网段
				ipRange = excludeIpRange(ipRange, cidrMerger.IpNetToRange(cidrMerger.StrToIpNet(hisExs)))
				// 排序、去重
				ipRange = cidrMerger.SortAndMerge(ipRange)

				// 转为网段形式字符串
				hosts = cidrMerger.IpNetToString(cidrMerger.IpRangeToIpNet(ipRange))
			}

			ctx.WriteString(strings.Join(hosts, ","))
		}
		apiParty.Get("/ip-line", getIpLine)
		apiParty.Get("/ip-line.txt", getIpLine)
		// 获取客户端IP
		apiParty.Get("/client", func(ctx iris.Context) {
			ctx.WriteString(ctx.RemoteAddr())
		})
	}

	return app
}

// 从网段中排除指定网段，比如排除私有地址或者指定地址
func excludeIpRange(irs []*cidrMerger.Range, exs []*cidrMerger.Range) []*cidrMerger.Range {
	irsTmp := make([]*cidrMerger.Range, 0, len(irs))
	for i, exRange := range exs {
		for _, irRange := range irs {
			if isIpBefore(irRange.Start, exRange.Start) {
				if isIpBefore(irRange.End, exRange.Start) {
					// ` |irRange.Start   |irRange.End
					// `                     |exRange.Start  |exRange.End
					irsTmp = append(irsTmp, irRange)
				} else if isIpBefore(irRange.End, exRange.End) {
					// ` |irRange.Start   |irRange.End
					// `    |exRange.Start  |exRange.End
					irsTmp = append(irsTmp, &cidrMerger.Range{Start: irRange.Start, End: ipMinusOne(exRange.Start)})
				} else {
					// ` |irRange.Start         |irRange.End
					// `    |exRange.Start  |exRange.End
					irsTmp = append(irsTmp, &cidrMerger.Range{Start: irRange.Start, End: ipMinusOne(exRange.Start)}, &cidrMerger.Range{Start: ipPlusOne(exRange.End), End: irRange.End})
				}
			} else {
				if isIpBefore(exRange.End, irRange.Start) {
					// `                    |irRange.Start   |irRange.End
					// ` |exRange.Start  |exRange.End
					irsTmp = append(irsTmp, irRange)
				} else if isIpBefore(irRange.End, exRange.End) {
					// `     |irRange.Start   |irRange.End
					// ` |exRange.Start          |exRange.End
					continue
				} else {
					// `     |irRange.Start   |irRange.End
					// ` |exRange.Start    |exRange.End
					irsTmp = append(irsTmp, &cidrMerger.Range{Start: ipPlusOne(exRange.Start), End: irRange.End})
				}
			}
		}

		// 下一轮需要用到上一轮的结果
		if i != len(exs)-1 {
			irs, irsTmp = irsTmp, irs[0:0]
		}
	}
	return irsTmp
}

func isIpBefore(a, b net.IP) bool {
	l := len(a)
	if l != len(b) {
		if len(a) == net.IPv6len {
			a = a.To4()
			if a == nil {
				panic("只支持同种类型的IP进行比较")
			}
		}
		if len(b) == net.IPv6len {
			b = b.To4()
			if b == nil {
				panic("只支持同种类型的IP进行比较")
			}
		}
	}
	for i := 0; i < l; i++ {
		if a[i] == b[i] {
			continue
		}
		return a[i] < b[i]
	}
	return false
}

// Ip+1
func ipPlusOne(ip net.IP) net.IP {
	l := len(ip)
	tmp := make(net.IP, l)
	copy(tmp, ip)

	for i := l - 1; i >= 0; i-- {
		if tmp[i] < 255 {
			tmp[i]++
			break
		} else {
			tmp[i] = 0
		}
	}

	return tmp
}

// Ip-1
func ipMinusOne(ip net.IP) net.IP {
	l := len(ip)
	tmp := make(net.IP, l)
	copy(tmp, ip)

	for i := l - 1; i >= 0; i-- {
		if tmp[i] > 0 {
			tmp[i]--
			break
		} else {
			tmp[i] = 255
		}
	}

	return tmp
}

// mergeIpV1 V1版本的合并规则，明确通过指定将达到指定数量的原始IP进行合并，
// 比如 'mask=16&level=10' 表示如果存在10个ip在16位掩码时的网络地址相同，则用16位掩码的网络地址表示它们。
// mask和level可以出现多次，且必须成对出现，当出现多次时，分别用他们和原始IP进行计算，并将每次得到的结果在最后进行合并。
// pri 参数表示默认情况下是否需要过滤掉内网IP
func mergeIpV1(hostIPNets []*net.IPNet, arr [][2]int) []*net.IPNet {
	dst := make([]*net.IPNet, 0, len(hostIPNets)*len(arr))
	for _, it := range arr {
		dst = append(dst, mergeIpByMaskAndLevel(hostIPNets, it[0], it[1])...)
	}
	return dst
}

// mergeIpV2 对IP进行简单合并，合并分为3步：
// 第1步，对原始数据进行合并，合并规则和版本1一样，但此步骤中明确指定mask为24，level值为第1个level参数的值
// 第2步，合并规则和第1步一样，但此步骤的输入数据不再是原始IP，而是第1步中生成的结果，且明确指定mask为16，level值为第2个level参数的值
// 第3步，再次重复第2步，此步骤的输入数据也不是原始IP，而是第2步中生成的结果，且明确指定mask为8，level值为第3个level参数的值。此步骤得到的结果为最终结果
func mergeIpV2(hostIPNets []*net.IPNet, level [3]int) []*net.IPNet {
	// 从第2步开始，需要把不符合进入下一步的数据筛选出来
	other := make([]*net.IPNet, 0, len(hostIPNets))

	// 第1步
	hostIPNets = mergeIpByMaskAndLevel(hostIPNets, 24, level[0])

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
	hostIPNets = mergeIpByMaskAndLevel(hostIPNets[:i], 16, level[1])

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
	hostIPNets = mergeIpByMaskAndLevel(hostIPNets[:i], 8, level[1])

	return append(hostIPNets, other...)
}

func mergeIpByMaskAndLevel(hosts []*net.IPNet, mask, level int) []*net.IPNet {
	l := len(hosts)

	ipNets := make([]*net.IPNet, l) // 目标网络地址
	j := 0
	for _, host := range hosts {
		size, _ := host.Mask.Size()
		if mask < size {
			mask := net.CIDRMask(mask, 8*net.IPv4len)
			tarIpNet := &net.IPNet{IP: host.IP.Mask(mask), Mask: mask}
			ipNets[j] = tarIpNet
			j++
		} else {
			ipNets[j] = host
			j++
		}
	}
	ipNets = ipNets[:j]
	ipNets = cidrMerger.IpRangeToIpNet(cidrMerger.SortAndMerge(cidrMerger.IpNetToRange(ipNets)))

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
