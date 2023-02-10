package pridns

import (
	"github.com/coredns/caddy"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name           string
		inputFileRules string
		want           *Config
		wantErr        bool
	}{
		{
			"正常配置",
			`pri-dns {
							adminPassword admin
							mysql {
								dataSourceName xx
							}
						}`,
			&Config{
				adminPassword: "admin",
				storeType:     storeTypeMySQL,
				mySQL:         MySQLConfig{dataSourceName: "xx"},
			},
			false,
		},
		{
			"存在多余指令",
			`pri-dns xx1 {
						}`,
			nil,
			true,
		},
		{
			"重复配置-插件",
			`pri-dns {
							adminPassword admin
							mysql {
								dataSourceName xx
							}
						}
						pri-dns {
						}`,
			nil,
			true,
		},
		{
			"重复配置-存储-mysql",
			`pri-dns {
							mysql {
								dataSourceName xx
							}
							mysql
						}`,
			nil,
			true,
		},
		{
			"重复配置-存储-etcd",
			`pri-dns {
							etcd {
							}
							etcd
						}`,
			nil,
			true,
		},
		{
			"重复配置-存储-redis",
			`pri-dns {
							redis {
							}
							redis
						}`,
			nil,
			true,
		},
		{
			"不支持的配置-root",
			`pri-dns {
							not_support v
						}`,
			nil,
			true,
		},
		{
			"不支持的配置-mysql",
			`pri-dns {
							mysql {
								not_support v
							}
						}`,
			nil,
			true,
		},
		{
			"不支持的配置-etcd",
			`pri-dns {
							etcd {
								not_support v
							}
						}`,
			nil,
			true,
		},
		{
			"不支持的配置-redis",
			`pri-dns {
							redis {
								not_support v
							}
						}`,
			nil,
			true,
		},
		{
			"没有配置存储",
			`pri-dns {
							adminPassword admin
						}`,
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := caddy.NewTestController("dns", tt.inputFileRules)
			got, err := parse(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse() got = %v, want %v", got, tt.want)
			}
		})
	}
}
