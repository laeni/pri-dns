package util

import "strings"

// GenAllMatchDomain 生成能够与目标域名匹配的所有解析域名的翻转格式
// 如 "a.b.example.com" -> ["com.example.b.a", "com.example.b.a.*", "com.example.b.*", "com.example.*"]
func GenAllMatchDomain(name string) []string {
	// 将域名按分隔符拆分为的切片
	domainSlice := strings.Split(name, ".")
	names := make([]string, 2, len(domainSlice)+2)
	names[0] = "*"
	names[1] = name

	for i := 0; i < len(domainSlice); i++ {
		names = append(names, "*."+strings.Join(domainSlice[i:], "."))
	}

	return names
}
