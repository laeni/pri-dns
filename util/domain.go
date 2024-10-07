package util

import "strings"

// ReverseDomainAndToSlice 翻转域名并解析为切片
// 如'a.example.com' -> ["com", "example", "a"]
func ReverseDomainAndToSlice(str string) []string {
	strArr := strings.Split(str, ".")
	l := len(strArr)
	for i := 0; i < l/2; i++ {
		strArr[l-1-i], strArr[i] = strArr[i], strArr[l-1-i]
	}

	if strArr[0] == "" {
		return strArr[1:]
	} else {
		return strArr
	}
}

// GenAllMatchDomain 生成能够与目标域名匹配的所有解析域名的翻转格式
// 如 "a.b.example.com" -> ["com.example.b.a", "com.example.b.a.*", "com.example.b.*", "com.example.*"]
func GenAllMatchDomain(name string) []string {
	// 将域名按分隔符拆分为的切片
	domainSlice := strings.Split(name, ".")
	domainSlice2 := ReverseDomainAndToSlice(name)
	names := make([]string, 3, len(domainSlice)*2+3)
	names[0] = "*"
	names[1] = name
	names[2] = strings.Join(domainSlice2, ".")

	for i := 0; i < len(domainSlice); i++ {
		names = append(names, "*."+strings.Join(domainSlice[i:], "."))
		names = append(names, strings.Join(domainSlice2[:i+1], ".")+".*")
	}

	return names
}
