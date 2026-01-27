// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package main

import (
	"flag"
	"fmt"

	"SkyeIM/app/message/api/internal/config"
	"SkyeIM/app/message/api/internal/handler"
	"SkyeIM/app/message/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/netx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/message-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	// 手动注册服务到 Etcd
	if len(c.Etcd.Hosts) > 0 {
		listenIP := c.Host
		if listenIP == "0.0.0.0" {
			listenIP = netx.InternalIp()
		}
		listenAddr := fmt.Sprintf("%s:%d", listenIP, c.Port)
		pub := discov.NewPublisher(c.Etcd.Hosts, c.Etcd.Key, listenAddr)
		defer pub.Stop()
		fmt.Printf("Registering service to Etcd: Key=%s, Addr=%s\n", c.Etcd.Key, listenAddr)
		// 启动注册
		if err := pub.KeepAlive(); err != nil {
			fmt.Printf("Failed to register service to Etcd: %v\n", err)
			panic(err)
		}
	}

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
