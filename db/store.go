package db

import (
	"context"
	"database/sql"
	"github.com/laeni/pri-dns/types"
)

type Store interface {
	// FindForwardByClient 查询客户端对应的转发配置，当 host 为 “” 时表示查询全局配置.
	FindForwardByClient(host string) []Forward

	// FindDomainByHostAndName 查询 qname 的解析记录。如果 host 不为空，则查询host下的解析，如果为空则只查询全局解析
	FindDomainByHostAndName(ctx context.Context, host, qname string) []Domain
}

// Domain 解析记录表.
type Domain struct {
	ID int64
	// 客户端地址（生效范围）。<br />如果全局生效，则该字段为空。
	Host string
	// 主机记录。由于可能存在泛域名，所以为了便于使用索引，存储时将采用反转格式，如：example.com
	Name string
	// 记录值
	Value string
	// TTL
	Ttl int32
	// 状态。<br />ENABLE-启用
	Status string
	// 记录类型。带"NO_"前缀的表示用于禁用全局解析。<br />A / NO_ALL / NO_A
	Type string
	// 优先级。值越小优先级越高。
	Priority int32
	// 创建时间
	CreateTime types.LocalTime
	// 修改时间
	UpdateTime types.LocalTime
}

// DomainSortPriority 仅仅用于根据 Priority 进行排序使用
type DomainSortPriority []Domain

func (a DomainSortPriority) Len() int {
	return len(a)
}

func (a DomainSortPriority) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a DomainSortPriority) Less(i, j int) bool {
	return a[i].Priority < a[j].Priority
}

// Forward 转发配置.
type Forward struct {
	ID int64
	// 创建时间
	CreateTime types.LocalTime
	// 修改时间
	UpdateTime types.LocalTime
	// 需要转发解析的域名
	Name string
	// 转发目标DNS服务器，可以是多个，多个以逗号分割
	Dns string
	// 生效范围。可选值：全局或某个具体Ip(如果为空则表示全局)
	Bind sql.NullString
	// 是否启用
	Enabled string
	// 该域名对应的解析历史。可能需要导出使用
	History sql.NullString
}
