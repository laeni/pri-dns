package pridns

import (
	"database/sql"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/laeni/pri-dns/db"
	"github.com/laeni/pri-dns/db/mysql"
)

func init() { plugin.Register("pri-dns", setup) }

func setup(c *caddy.Controller) error {
	config, err := parse(c)
	if err != nil {
		return err
	}

	store, err := initDb(config)
	if err != nil {
		return err
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return &PriDns{
			Config: config,
			Next:   next,
			Store:  store,
		}
	})

	return nil
}

func parse(c *caddy.Controller) (*Config, error) {
	config := &Config{}

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
					config.adminPassword = adminPasswordArgs[0]
				case "mysql":
					if config.storeType != "" {
						return nil, c.Err("配置重复定义: mysql")
					}
					config.storeType = storeTypeMySQL

					for c.NextBlock() {
						switch c.Val() {
						case "dataSourceName":
							dataSourceNameArgs := c.RemainingArgs()
							if len(dataSourceNameArgs) != 1 {
								return nil, c.Errf("dataSourceName 配置错误")
							}
							config.mySQL.dataSourceName = dataSourceNameArgs[0]
						default:
							return nil, c.Errf("不支持的配置: %s", c.Val())
						}
					}
				case "etcd":
					if config.storeType != "" {
						return nil, c.Err("配置重复定义: etcd")
					}
					config.storeType = storeTypeEtcd

					for c.NextBlock() {
						switch c.Val() {
						default:
							return nil, c.Errf("不支持的配置: %s", c.Val())
						}
					}
				case "redis":
					if config.storeType != "" {
						return nil, c.Err("配置重复定义: redis")
					}
					config.storeType = storeTypeRedis

					for c.NextBlock() {
						switch c.Val() {
						default:
							return nil, c.Errf("不支持的配置: %s", c.Val())
						}
					}
				default:
					return nil, c.Errf("不支持的配置: %s", c.Val())
				}
			}
		}
	}

	if config.storeType == "" {
		return nil, c.Errf("必须至少使用其中一种存储")
	}

	return config, nil
}

func initDb(config *Config) (db.Store, error) {
	switch config.storeType {
	case storeTypeMySQL:
		d, err := sql.Open("mysql", config.mySQL.dataSourceName)
		if err != nil {
			log.Fatal(err)
		}
		caddy.OnProcessExit = append(caddy.OnProcessExit, func() {
			_ = d.Close()
		})
		store := mysql.NewStore(d)
		return &store, nil
	}
	return nil, nil
}
