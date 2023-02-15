package cidr_merger

import (
	"net"
	"strings"
)

func parseIp(str string) net.IP {
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
func parse(text string) (IRange, error) {
	if index := strings.IndexByte(text, '/'); index != -1 {
		if _, network, err := net.ParseCIDR(text); err == nil {
			return IpNetWrapper{network}, nil
		} else {
			return nil, err
		}
	}
	if ip := parseIp(text); ip != nil {
		return IpWrapper{ip}, nil
	}
	if index := strings.IndexByte(text, '-'); index != -1 {
		if start, end := parseIp(text[:index]), parseIp(text[index+1:]); start != nil && end != nil {
			if len(start) == len(end) && !lessThan(end, start) {
				return &Range{start: start, end: end}, nil
			}
		}
		return nil, &net.ParseError{Type: "range", Text: text}
	}
	return nil, &net.ParseError{Type: "ip/CIDR address/range", Text: text}
}

// MergeIp 对ip进行合并
// see https://github.com/zhanhb/cidr-merger
func MergeIp(in []string, typeRange bool) []string {
	var result []IRange
	for _, text := range in {
		if maybe, err := parse(text); err != nil {
			continue
		} else {
			result = append(result, maybe)
		}
	}

	result = sortAndMerge(result)
	result = convertBatch(result, typeRange)

	resArr := make([]string, 0)
	for _, r := range result {
		resArr = append(resArr, r.String())
	}

	return resArr
}
