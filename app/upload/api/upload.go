package main

import (
	"flag"
	"fmt"

	"SkyeIM/app/upload/api/internal/config"
	"SkyeIM/app/upload/api/internal/handler"
	"SkyeIM/app/upload/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/upload-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting upload-api server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
