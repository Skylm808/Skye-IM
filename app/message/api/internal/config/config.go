// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	MessageRpc zrpc.RpcClientConf
	GroupRpc   zrpc.RpcClientConf `json:",optional"`
	FriendRpc  zrpc.RpcClientConf `json:",optional"`
}
