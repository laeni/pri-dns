package util

import "strings"

// ReverseDomainAndToSlice 翻转域名并解析为切片
// 目前已知多后缀域名只有一个"com.cn"，所以处理该类型的域名时会将"com.cn"看为一个整体
// 如'a.example.com' -> ["com", "example", "a"]
// 如'a.example.com.cn' -> ["com.cn", "example", "a"]
func ReverseDomainAndToSlice(str string) []string {
	strArr := strings.Split(str, ".")
	l := len(strArr)
	for i := 0; i < l/2; i++ {
		strArr[l-1-i], strArr[i] = strArr[i], strArr[l-1-i]
	}

	// 处理多后缀域名
	if len(strArr) >= 3 && strArr[0] == "cn" && strArr[1] == "com" {
		strArr[0] = "com.cn"
		strArr = SliceRemoveItem(strArr, 1)
	}

	if strArr[0] == "" {
		return strArr[1:]
	} else {
		return strArr
	}
}

// ReverseDomain 翻转域名
// 如'a.example.com' -> 'com.example.a'
func ReverseDomain(str string) string {
	return strings.Join(ReverseDomainAndToSlice(str), ".")
}

// GenAllMatchDomain 生成能够与目标域名匹配的所有解析域名的翻转格式
// 如 "a.b.example.com" -> ["com.example.b.a", "com.example.b.*", "com.example.*"]
func GenAllMatchDomain(name string) []string {
	// 反转域名并得到各部分组成的切片
	domainSlice := ReverseDomainAndToSlice(name)
	names := []string{strings.Join(domainSlice, ".")}

	if len(domainSlice) > 2 {
		for i := len(domainSlice) - 2; i >= 1; i-- {
			names = append(names, strings.Join(domainSlice[:i+1], ".")+".*")
		}
	}

	return names
}
