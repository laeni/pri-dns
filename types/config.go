package types

import (
	"crypto/tls"
	"time"
)

// Config 表示插件配置
type Config struct {
	AdminPassword string
	StoreType     string
	MySQL         MySQLConfig
	Tls           map[string]*tls.Config // TLS 配置。key 为IP，value 为该IP对应的主机名与 TLS 配置
	HealthCheck   HealthCheckConfig      // 健康检查配置
}

type MySQLConfig struct {
	// DataSourceName 为MySQL数据库连接地址
	DataSourceName string
}

// HealthCheckConfig 为健康检查配置，配置时格式与 forward 插件配置相同
type HealthCheckConfig struct {
	HcInterval         time.Duration
	HcRecursionDesired bool
	HcDomain           string
}
