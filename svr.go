package pridns

import (
	"context"
	"fmt"
	"github.com/kataras/iris/v12"
	context2 "github.com/kataras/iris/v12/context"
	"github.com/laeni/pri-dns/db"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	app *iris.Application // 应用程序实例
)

// StartApp 启动服务。启动时需要 PriDns 实例来注册关闭回调钩子以及读取相关配置
func StartApp(p *PriDns) error {
	// 目前不支持修改应用配置
	if app != nil {
		return nil
	}
	app = newApp(p.Store)

	var startError error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		startError = func() error {
			var err error

			// 检测服务是否正常启动
			go func() {
				var healthErr error
				for i := 0; i < 100; i++ {
					if err != nil {
						wg.Done()
						return
					}
					resp, errRes := http.Get(fmt.Sprintf("http://127.0.0.1:%s/health", strings.Split(p.Config.ServerPort, ":")[1]))
					if errRes == nil && (resp.StatusCode >= 200 && resp.StatusCode < 300) {
						wg.Done()
						return
					}
					healthErr = errRes
					time.Sleep(100 * time.Millisecond)
				}
				log.Error("后台服务健康检查无法通过", healthErr)
			}()

			err = app.Listen(p.Config.ServerPort)
			return err
		}()
	}()
	wg.Wait()
	return startError
}

func newApp(store db.Store) *iris.Application {
	app = iris.New()
	app.Get("/health", func(c context2.Context) {
		_, _ = c.WriteString("OK")
	})

	apiParty := app.Party("/api")
	{
		getIpLine := func(ctx context2.Context) {
			level := ctx.URLParamIntDefault("level", 0)

			addr := ctx.RemoteAddr()
			hosts := store.FindHistoryByHost(context.Background(), addr)
			ctx.WriteString(fmt.Sprintf("Level: %d", level) + " - " + strings.Join(hosts, ","))
		}
		apiParty.Get("/ip-line", getIpLine)
		apiParty.Get("/ip-line.txt", getIpLine)
	}

	return app
}
