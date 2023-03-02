package types

import (
	"crypto/tls"
	"time"
)

// Config 表示插件配置
type Config struct {
	HostPort      string // 控制台管理端口,如果不配置则不会开启控制台
	AdminPassword string // 管理员身份确认密码，提供该密码可以拥有管理员身份，否则将被视为普通用户
	StoreType     string
	MySQL         MySQLConfig
	Tls           map[string]*tls.Config // TLS 配置。key 为IP，value 为该IP对应的主机名与 TLS 配置
	HealthCheck   HealthCheckConfig      // 健康检查配置
}

type MySQLConfig struct {
	DataSourceName  string        // MySQL数据库连接地址
	MaxIdleConns    int           // 空闲连接池中连接的最大数量（默认：2）
	MaxOpenConns    int           // 打开数据库连接的最大数量(默认：0,无限制)
	ConnMaxLifetime time.Duration // 连接可复用的最大时间
}

// HealthCheckConfig 为健康检查配置，配置时格式与 forward 插件配置相同
type HealthCheckConfig struct {
	HcInterval         time.Duration
	HcRecursionDesired bool
	HcDomain           string
}
