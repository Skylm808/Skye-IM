package main

import (
	"flag"
	"fmt"
	"net/http"

	"SkyeIM/app/ws/internal/config"
	"SkyeIM/app/ws/internal/conn"
	"SkyeIM/app/ws/internal/handler"
	"SkyeIM/app/ws/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/ws.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建服务上下文
	ctx := svc.NewServiceContext(c)

	// 创建 Hub（传入 ServiceContext）
	hub := conn.NewHub(ctx)
	go hub.Run()

	// 创建 HTTP 服务器
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 创建 WebSocket 处理器
	wsHandler := handler.NewWsHandler(ctx, hub)

	// 注册 WebSocket 路由
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/ws",
		Handler: wsHandler.ServeHTTP,
	})

	// 健康检查接口
	server.AddRoute(rest.Route{
		Method: http.MethodGet,
		Path:   "/health",
		Handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(`{"status":"ok","online":%d}`, hub.OnlineCount())))
		},
	})

	fmt.Printf("Starting WebSocket server at %s:%d...\n", c.Host, c.Port)
	logx.Infof("WebSocket server listening on %s:%d", c.Host, c.Port)
	server.Start()
}
