package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	MySQL struct {
		DataSource string
	}
	Cache cache.CacheConf

	// WebSocket 服务地址
	WsServiceUrl string

	// WebSocket 内部推送鉴权（可选）
	WsPushSecret string `json:",optional"`
}
