package db

import (
	"github.com/laeni/pri-dns/types"
)

type Store interface {
	// FindForwardByHostAndName 查询客户端对应的转发配置，当 host 为 “” 时表示查询全局配置.
	FindForwardByHostAndName(host, name string) []Forward

	// FindDomainByHostAndName 查询 qname 的解析记录。如果 host 不为空，则查询host下的解析，如果为空则只查询全局解析
	FindDomainByHostAndName(host, qname string) []Domain

	// SavaHistory 保存历史
	SavaHistory(name string, newHis []string) error

	// FindHistoryByHost 查询客户端对应的解析历史，当 host 为 “” 时表示查询全局配置.
	// 其中返回值的二个值表示需要排除的网段
	FindHistoryByHost(host string) ([]string, []string)
}

type RecordFilter interface {
	ClientHostVal() string
	NameVal() string
	DenyGlobalVal() bool
}

// Domain 解析记录表.
type Domain struct {
	ID         int64
	ClientHost string          // 客户端地址（生效范围）。<br />如果全局生效，则该字段为空。
	Name       string          // 主机记录。由于可能存在泛域名，所以为了便于使用索引，存储时将采用反转格式，如：example.com
	Value      string          // 记录值
	Ttl        int32           // TTL
	DnsType    string          // 记录类型。<br />A | AAAA
	DenyGlobal bool            // 是否拒绝全局解析
	Enable     bool            // 是否启用
	CreateTime types.LocalTime // 创建时间
	UpdateTime types.LocalTime // 修改时间
}

func (d Domain) ClientHostVal() string {
	return d.ClientHost
}
func (d Domain) NameVal() string {
	return d.Name
}
func (d Domain) DenyGlobalVal() bool {
	return d.DenyGlobal
}

// Forward 转发配置.
type Forward struct {
	ID         int64
	ClientHost string          // 客户端地址（生效范围）。<br />如果全局生效，则该字段为空。
	Name       string          // 需要转发解析的域名
	DnsSvr     []string        // 转发目标DNS服务器
	DenyGlobal bool            // 是否拒绝全局解析
	Enable     bool            // 是否启用
	CreateTime types.LocalTime // 创建时间
	UpdateTime types.LocalTime // 修改时间
}

func (f Forward) ClientHostVal() string {
	return f.ClientHost
}
func (f Forward) NameVal() string {
	return f.Name
}
func (f Forward) DenyGlobalVal() bool {
	return f.DenyGlobal
}

// History 转发解析历史.
type History struct {
	ID      int64
	Name    string   // 需要转发解析的域名
	History []string // 解析记录，用于导出使用
}
