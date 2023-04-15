package cidr_merger

import (
	"fmt"
	"net"
	"strings"
)

func ParseIp(str string) net.IP {
	for _, b := range str {
		switch b {
		case '.':
			return net.ParseIP(str).To4()
		case ':':
			return net.ParseIP(str).To16()
		}
	}
	return nil
}

// maybe IpWrapper, Range or IpNetWrapper is returned
// 128.0.0.0-255.255.255.255 -> Range
// 192.168.9.0/22            -> IpNetWrapper
// 127.0.0.1                 -> IpWrapper
func Parse(text string) (IRange, error) {
	if strings.IndexByte(text, '/') != -1 {
		if _, network, err := net.ParseCIDR(text); err == nil {
			return IpNetWrapper{network}, nil
		} else {
			return nil, err
		}
	}
	if index := strings.IndexByte(text, '-'); index != -1 {
		if start, end := ParseIp(text[:index]), ParseIp(text[index+1:]); start != nil && end != nil {
			if len(start) == len(end) && !lessThan(end, start) {
				return &Range{Start: start, End: end}, nil
			}
		}
		return nil, &net.ParseError{Type: "range", Text: text}
	}
	if ip := ParseIp(text); ip != nil {
		return IpWrapper{ip}, nil
	}
	return nil, &net.ParseError{Type: "ip/CIDR address/range", Text: text}
}

// StrToIpNet 将String格式的IP转换为网段对象
func StrToIpNet(hosts []string) []*net.IPNet {
	hostIPNets := make([]*net.IPNet, len(hosts))
	i := 0
	for _, host := range hosts {
		// 如果不是 CIDR 格式则在末尾拼接掩码将其转换为 CIDR 格式
		if strings.IndexByte(host, '/') == -1 {
			host = fmt.Sprintf("%s/%d", host, 32)
		}

		_, ipNet, err := net.ParseCIDR(host)
		if err != nil {
			continue
		}
		hostIPNets[i] = ipNet
		i++
	}
	return hostIPNets[:i]
}

// IpNetToString 将网段对象格式的IP转换为String
func IpNetToString(ipNets []*net.IPNet) []string {
	ips := make([]string, len(ipNets))
	for i, ipNet := range ipNets {
		ips[i] = ipNet.String()
	}
	return ips
}

func IpNetToIRange(in []*net.IPNet) []IRange {
	result := make([]IRange, len(in))
	for i, it := range in {
		result[i] = IpNetWrapper{IPNet: it}
	}
	return result
}

func IpNetToRange(in []*net.IPNet) []*Range {
	wrappers := IpNetToIRange(in)
	ranges := make([]*Range, 0, len(wrappers))
	for _, e := range wrappers {
		ranges = append(ranges, e.ToRange())
	}
	return ranges
}

func IpRangeToIpNet(irs []*Range) []*net.IPNet {
	resArr := make([]*net.IPNet, 0, len(irs))
	for _, r := range irs {
		resArr = append(resArr, r.ToIpNets()...)
	}
	return resArr
}
