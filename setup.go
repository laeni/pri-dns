package pridns

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() { plugin.Register("pridns", setup) }

func setup(c *caddy.Controller) error {
	config, err := parse(c)
	if err != nil {
		return err
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return &PriDns{
			Config: config,
			Next:   next,
		}
	})

	return nil
}

func parse(c *caddy.Controller) (*Config, error) {
	config := &Config{}

	// 解析
	i := 0
	for c.Next() {
		// 同一个插件链只允许定义一次
		if i > 0 {
			return nil, plugin.ErrOnce
		}
		i++

		_ = c.RemainingArgs()

		for c.NextBlock() {
			switch c.Val() {
			case "dataSourceName":
				dataSourceNameArgs := c.RemainingArgs()
				if len(dataSourceNameArgs) != 1 {
					return nil, c.Errf("dataSourceName 配置错误")
				}
				config.DataSourceName = dataSourceNameArgs[0]
			default:
			}
		}
	}

	if config.DataSourceName == "" {
		return nil, c.Errf("必要不能配置为空")
	}

	return config, nil
}
