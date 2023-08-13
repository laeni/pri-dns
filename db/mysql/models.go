package mysql

import (
	"database/sql"

	types "github.com/laeni/pri-dns/types"
)

// 解析记录表.
type Domain struct {
	ID int64 `gorm:"primaryKey"`
	// 客户端地址（生效范围）。<br />如果全局生效，则该字段为空。
	ClientHost string
	// 主机记录。由于可能存在泛域名，所以为了便于使用索引，存储时将采用反转格式，如：example.com
	Name string
	// 记录值
	Value sql.NullString
	// TTL
	Ttl sql.NullInt32
	// 记录类型。<br />A | AAAA
	DnsType sql.NullString
	// 是否拒绝全局解析
	DenyGlobal string
	// 状态。<br />ENABLE-启用
	Status string
	// 创建时间
	CreateTime types.LocalTime
	// 修改时间
	UpdateTime types.LocalTime
}

func (Domain) TableName() string {
	return "domain"
}

// 转发配置.
type Forward struct {
	ID int64 `gorm:"primaryKey"`
	// 客户端地址（生效范围）。<br />如果全局生效，则该字段为空。
	ClientHost string
	// 需要转发解析的域名
	Name string
	// 转发目标DNS服务器，可以是多个，多个以逗号分割
	DnsSvr sql.NullString
	// 是否拒绝全局解析
	DenyGlobal string
	// 状态。<br />ENABLE-启用
	Status string
	// 创建时间
	CreateTime types.LocalTime
	// 修改时间
	UpdateTime types.LocalTime
}

func (Forward) TableName() string {
	return "forward"
}

// 解析历史.
type History struct {
	ID int64 `gorm:"primaryKey"`
	// 需要转发解析的域名
	Name string
	// 解析记录，用于导出使用，多个以逗号分割
	History sql.NullString
	// 创建时间
	CreateTime types.LocalTime
	// 修改时间
	UpdateTime types.LocalTime
}

func (History) TableName() string {
	return "history"
}

// 用于排除历史中的网段。由于历史数据需要按照一定规则聚合精简，所以处理后的范围可能包含一些特殊网段，比如内网等，所以这里列出的数据将在处理后生效，即精确排除这里列出的网段.
type HistoryEx struct {
	ID int64 `gorm:"primaryKey"`
	// 客户端地址（生效范围）。<br />如果全局生效，则该字段为空。
	ClientHost string
	// 需要排除的网段
	IpNet string
	// 是否拒绝全局.为了简化，和 domain 表一样当 clent_host 为空时的记录对所有人生效，但是特定的某个而可以排除这种默认设置
	DenyGlobal string
	// 标签/分组
	Label sql.NullString
	// 创建时间
	CreateTime types.LocalTime
	// 修改时间
	UpdateTime types.LocalTime
}

func (HistoryEx) TableName() string {
	return "history_ex"
}
