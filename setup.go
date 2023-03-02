package pridns

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	pkgtls "github.com/coredns/coredns/plugin/pkg/tls"
	_ "github.com/go-sql-driver/mysql"
	"github.com/laeni/pri-dns/db"
	"github.com/laeni/pri-dns/db/mysql"
	"github.com/laeni/pri-dns/forward"
	"github.com/laeni/pri-dns/types"
	"github.com/miekg/dns"
	"strconv"
	"time"
)

func init() { plugin.Register("pri-dns", setup) }

func setup(c *caddy.Controller) error {
	config, err := parsePriDns(c)
	if err != nil {
		return err
	}

	store, err := initDb(c, config)
	if err != nil {
		return err
	}

	p := NewPriDns(config, store)
	c.OnStartup(p.initFunc)
	c.OnShutdown(p.closeFunc)

	if config.HostPort != "" {
		err := StartApp(p)
		if err != nil {
			return err
		}
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		p.Next = next
		return p
	})

	return nil
}

func parsePriDns(c *caddy.Controller) (*types.Config, error) {
	config := &types.Config{
		Tls:         make(map[string]*tls.Config),
		HealthCheck: types.HealthCheckConfig{HcInterval: 5000 * time.Millisecond, HcDomain: "."},
		MySQL:       types.MySQLConfig{ConnMaxLifetime: 10 * time.Minute},
	}

	// 解析
	for i := 1; c.Next(); i++ {
		// 同一个插件链只允许定义一次
		if i > 1 {
			return nil, plugin.ErrOnce
		}

		// 目前不需要指令
		args := c.RemainingArgs()
		if len(args) != 0 {
			return nil, c.Errf("Wrong argument count or unexpected line ending after '%s'", args[0])
		}

		// 进入到配置块中（由于 caddyfile.Dispenser 不支持嵌套块，所以这里不能使用 NextBlock()）
		if c.Next() && c.Val() == "{" {
			// 循环解析大块中的每一项配置（注意：这里的'}'为大块结束，与上一行‘if’中的'{'对应，大块中的小块需确保在一次解析中处理完成）
			for c.Next() && c.Val() != "}" {
				switch c.Val() {
				case "adminPassword":
					adminPasswordArgs := c.RemainingArgs()
					if len(adminPasswordArgs) != 1 {
						return nil, c.Err("'adminPassword' 配置错误，它有且仅有一个参数")
					}
					config.AdminPassword = adminPasswordArgs[0]
				case "adminPort":
					adminPortArgs := c.RemainingArgs()
					if len(adminPortArgs) != 1 {
						return nil, c.Err("'adminPort' 配置错误，它有且仅有一个参数")
					}
					config.HostPort = adminPortArgs[0]
				case "mysql":
					if config.StoreType != "" {
						return nil, c.Err("配置重复定义: mysql")
					}
					config.StoreType = storeTypeMySQL

					for c.NextBlock() {
						switch c.Val() {
						case "dataSourceName":
							dataSourceNameArgs := c.RemainingArgs()
							if len(dataSourceNameArgs) != 1 {
								return nil, c.Errf("dataSourceName 配置错误")
							}
							config.MySQL.DataSourceName = dataSourceNameArgs[0]
						case "maxIdleConns":
							args := c.RemainingArgs()
							if len(args) != 1 {
								return nil, fmt.Errorf("maxIdleConns 参数个数有误")
							}
							maxIdleConns, err := strconv.Atoi(args[0])
							if err != nil {
								return nil, err
							}
							config.MySQL.MaxIdleConns = maxIdleConns
						case "maxOpenConns":
							args := c.RemainingArgs()
							if len(args) != 1 {
								return nil, fmt.Errorf("maxOpenConns 参数个数有误")
							}
							maxOpenConns, err := strconv.Atoi(args[0])
							if err != nil {
								return nil, err
							}
							config.MySQL.MaxOpenConns = maxOpenConns
						case "connMaxLifetime":
							args := c.RemainingArgs()
							if len(args) != 1 {
								return nil, fmt.Errorf("connMaxLifetime 参数个数有误")
							}
							dur, err := time.ParseDuration(args[0])
							if err != nil {
								return nil, err
							}
							if dur < 0 {
								return nil, fmt.Errorf("connMaxLifetime can't be negative: %d", dur)
							}
							config.MySQL.ConnMaxLifetime = dur
						default:
							return nil, c.Errf("不支持的配置: %s", c.Val())
						}
					}
				case "etcd":
					if config.StoreType != "" {
						return nil, c.Err("配置重复定义: etcd")
					}
					config.StoreType = storeTypeEtcd

					for c.NextBlock() {
						switch c.Val() {
						default:
							return nil, c.Errf("不支持的配置: %s", c.Val())
						}
					}
				case "file":
					if config.StoreType != "" {
						return nil, c.Err("配置重复定义: file")
					}
					config.StoreType = storeTypeRedis

					for c.NextBlock() {
						switch c.Val() {
						default:
							return nil, c.Errf("不支持的配置: %s", c.Val())
						}
					}
				case "tls":
					// tls 后面不能有其他配置
					if len(c.RemainingArgs()) > 0 {
						return nil, c.ArgErr()
					}

					var servername string
					var hosts []string
					var tlsConfig *tls.Config
					for c.NextBlock() {
						switch c.Val() {
						case "cert":
							if tlsConfig != nil {
								return nil, c.Err("配置重复定义: cert")
							}
							// 解析证书
							args := c.RemainingArgs()
							if len(args) > 3 {
								return nil, c.ArgErr()
							}
							tlsConfigTmp, err := pkgtls.NewTLSConfigFromArgs(args...)
							if err != nil {
								return nil, err
							}
							tlsConfig = tlsConfigTmp
						case "servername":
							if servername != "" {
								return nil, c.Err("配置重复定义: servername")
							}
							if !c.NextArg() {
								return nil, c.ArgErr()
							}
							servername = c.Val()
						case "hosts":
							if hosts != nil {
								return nil, c.Err("配置重复定义: hosts")
							}
							hosts = c.RemainingArgs()
							if len(hosts) == 0 {
								return nil, c.ArgErr()
							}
						default:
							return nil, c.Errf("unknown policy '%s'", c.Val())
						}
					}
					if len(hosts) == 0 {
						return nil, c.Err("tls 配置缺失")
					}
					if tlsConfig == nil {
						tlsConfig = new(tls.Config)
					}
					if servername != "" {
						tlsConfig.ServerName = servername
					}
					tlsConfig.ClientSessionCache = forward.ClientSessionCache

					for _, host := range hosts {
						if it, ok := config.Tls[host]; ok {
							if it != nil {
								return nil, c.Errf("配置冲突! host: %s", host)
							}
						}
						config.Tls[host] = tlsConfig
					}
				case "health_check":
					if !c.NextArg() {
						return nil, c.ArgErr()
					}
					dur, err := time.ParseDuration(c.Val())
					if err != nil {
						return nil, err
					}
					if dur < 0 {
						return nil, fmt.Errorf("health_check can't be negative: %d", dur)
					}
					config.HealthCheck.HcInterval = dur
					config.HealthCheck.HcDomain = "."

					for c.NextArg() {
						switch hcOpts := c.Val(); hcOpts {
						case "no_rec":
							config.HealthCheck.HcRecursionDesired = false
						case "domain":
							if !c.NextArg() {
								return nil, c.ArgErr()
							}
							hcDomain := c.Val()
							if _, ok := dns.IsDomainName(hcDomain); !ok {
								return nil, fmt.Errorf("health_check: invalid domain name %s", hcDomain)
							}
							config.HealthCheck.HcDomain = plugin.Name(hcDomain).Normalize()
						default:
							return nil, fmt.Errorf("health_check: unknown option %s", hcOpts)
						}
					}
				default:
					return nil, c.Errf("不支持的配置: %s", c.Val())
				}
			}
		}
	}

	if config.StoreType == "" {
		return nil, c.Errf("必须至少使用其中一种存储")
	}

	return config, nil
}

func initDb(c *caddy.Controller, config *types.Config) (db.Store, error) {
	switch config.StoreType {
	case storeTypeMySQL:
		d, err := sql.Open("mysql", config.MySQL.DataSourceName)
		if err != nil {
			log.Fatal(err)
		}
		c.OnShutdown(func() error {
			return d.Close()
		})
		// SetMaxIdleConns 设置空闲连接池中连接的最大数量（默认：2）
		d.SetMaxIdleConns(config.MySQL.MaxIdleConns)
		// SetMaxOpenConns 设置打开数据库连接的最大数量(默认：0,无限制)
		d.SetMaxOpenConns(config.MySQL.MaxOpenConns)
		// SetConnMaxLifetime 设置了连接可复用的最大时间。
		d.SetConnMaxLifetime(config.MySQL.ConnMaxLifetime)
		store := mysql.NewStore(d)
		return &store, nil
	}
	return nil, nil
}
