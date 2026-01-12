package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	// JWT 配置
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}

	// Message RPC 配置
	MessageRpc zrpc.RpcClientConf

	// Friend RPC 配置
	FriendRpc zrpc.RpcClientConf

	// Group RPC 配置
	GroupRpc zrpc.RpcClientConf

	// WebSocket 配置
	WebSocket struct {
		PingInterval   int   // 心跳间隔（秒）
		PongTimeout    int   // pong 超时（秒）
		MaxMessageSize int64 // 最大消息大小（字节）
	}

	// Redis 配置
	Redis struct {
		Host string
		Type string
		Pass string
	}

	// 内部推送接口鉴权（可选）
	PushEvent struct {
		Secret string `json:",optional"`
	} `json:",optional"`
}
