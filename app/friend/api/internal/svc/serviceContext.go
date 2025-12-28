// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"SkyeIM/app/friend/api/internal/config"
	"SkyeIM/app/friend/rpc/friendclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	FriendRpc friendclient.Friend
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:    c,
		FriendRpc: friendclient.NewFriend(zrpc.MustNewClient(c.FriendRpc)),
	}
}
