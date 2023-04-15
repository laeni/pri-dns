package cidr_merger

import (
	"bytes"
	"fmt"
	"math/bits"
	"net"
	"sort"
	"strconv"
)

func ipToString(ip net.IP) string {
	if len(ip) == net.IPv6len {
		if ipv4 := ip.To4(); len(ipv4) == net.IPv4len {
			return "::ffff:" + ipv4.String()
		}
	}
	return ip.String()
}

type IRange interface {
	ToIpNets() []*net.IPNet
	ToRange() *Range
	String() string
}

type Range struct {
	Start net.IP
	End   net.IP
}

func (r *Range) ToIpNets() []*net.IPNet {
	start, end := r.Start, r.End
	ipBits := len(start) * 8
	if ipBits != len(end)*8 {
		if len(start) == net.IPv6len {
			start = start.To4()
			if start == nil {
				assert(false, "len(r.Start) == len(r.End)")
			}
		}
		if len(end) == net.IPv6len {
			end = end.To4()
			if end == nil {
				assert(false, "len(r.Start) == len(r.End)")
			}
		}
	}
	var result []*net.IPNet
	for {
		assert(bytes.Compare(start, end) <= 0, "start <= End")
		cidr := max(prefixLength(xor(addOne(end), start)), ipBits-trailingZeros(start))
		ipNet := &net.IPNet{IP: start, Mask: net.CIDRMask(cidr, ipBits)}
		result = append(result, ipNet)
		tmp := lastIp(ipNet)
		if !lessThan(tmp, end) {
			return result
		}
		start = addOne(tmp)
	}
}
func (r *Range) ToRange() *Range {
	return r
}
func (r *Range) String() string {
	return ipToString(r.Start) + "-" + ipToString(r.End)
}

type IpWrapper struct {
	net.IP
}

func (r IpWrapper) ToIpNets() []*net.IPNet {
	ipBits := len(r.IP) * 8
	return []*net.IPNet{
		{IP: r.IP, Mask: net.CIDRMask(ipBits, ipBits)},
	}
}
func (r IpWrapper) ToRange() *Range {
	return &Range{Start: r.IP, End: r.IP}
}
func (r IpWrapper) String() string {
	return ipToString(r.IP)
}

type IpNetWrapper struct {
	*net.IPNet
}

func (r IpNetWrapper) ToIpNets() []*net.IPNet {
	return []*net.IPNet{r.IPNet}
}
func (r IpNetWrapper) ToRange() *Range {
	ipNet := r.IPNet
	return &Range{Start: ipNet.IP, End: lastIp(ipNet)}
}
func (r IpNetWrapper) String() string {
	ip, mask := r.IP, r.Mask
	if ones, bitCount := mask.Size(); bitCount != 0 {
		return ipToString(ip) + "/" + strconv.Itoa(ones)
	}
	return ipToString(ip) + "/" + mask.String()
}

func lessThan(a, b net.IP) bool {
	if lenA, lenB := len(a), len(b); lenA != lenB {
		return lenA < lenB
	}
	return bytes.Compare(a, b) < 0
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

func allFF(ip []byte) bool {
	for _, c := range ip {
		if c != 0xff {
			return false
		}
	}
	return true
}

func prefixLength(ip net.IP) int {
	for index, c := range ip {
		if c != 0 {
			return index*8 + bits.LeadingZeros8(c) + 1
		}
	}
	// special case for overflow
	return 0
}

func trailingZeros(ip net.IP) int {
	ipLen := len(ip)
	for i := ipLen - 1; i >= 0; i-- {
		if c := ip[i]; c != 0 {
			return (ipLen-i-1)*8 + bits.TrailingZeros8(c)
		}
	}
	return ipLen * 8
}

// 获取网段的最后一个Ip
func lastIp(ipNet *net.IPNet) net.IP {
	ip, mask := ipNet.IP, ipNet.Mask
	maskLen := len(mask)
	ipLen := len(ip)
	if ipLen != maskLen {
		ip = ip.To4()
		if ip != nil {
			ipLen = len(ip)
			if maskLen != ipLen {
				mask = mask[12:16]
				maskLen = len(mask)
			}
		} else {
			assert(false, "unexpected IPNet %v", ipNet)
		}
	}
	res := make(net.IP, ipLen)
	for i := 0; i < ipLen; i++ {
		res[i] = ip[i] | ^mask[i]
	}
	return res
}

// 将最右边第一个不是 255 的的数 +1
func addOne(ip net.IP) net.IP {
	ipLen := len(ip)
	res := make(net.IP, ipLen)
	for i := ipLen - 1; i >= 0; i-- {
		if t := ip[i]; t != 0xFF {
			res[i] = t + 1
			copy(res, ip[0:i])
			break
		}
	}
	return res
}

// return a ^ b
func xor(a, b net.IP) net.IP {
	ipLen := len(a)
	assert(ipLen == len(b), "a=%v, b=%v", a, b)
	res := make(net.IP, ipLen)
	for i := ipLen - 1; i >= 0; i-- {
		res[i] = a[i] ^ b[i]
	}
	return res
}

type Ranges []*Range

func (s Ranges) Len() int { return len(s) }
func (s Ranges) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s Ranges) Less(i, j int) bool {
	return lessThan(s[i].Start, s[j].Start)
}

// SortAndMerge 排序并合并
func SortAndMerge(ranges []*Range) []*Range {
	if len(ranges) < 2 {
		return ranges
	}
	sort.Sort(Ranges(ranges))

	res := make([]*Range, 0, len(ranges))
	now := ranges[0]
	familyLength := len(now.Start)
	start, end := now.Start, now.End
	for i, count := 1, len(ranges); i < count; i++ {
		item := ranges[i]
		if fl := len(item.Start); fl != familyLength {
			res = append(res, &Range{start, end})
			familyLength = fl
			start, end = item.Start, item.End
			continue
		}
		if allFF(end) || !lessThan(addOne(end), item.Start) {
			if lessThan(end, item.End) {
				end = item.End
			}
		} else {
			res = append(res, &Range{start, end})
			start, end = item.Start, item.End
		}
	}
	return append(res, &Range{start, end})
}

func assert(condition bool, format string, args ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assert failed: "+format, args...))
	}
}
