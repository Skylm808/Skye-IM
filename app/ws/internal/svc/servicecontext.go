package svc

import (
	"SkyeIM/app/friend/rpc/friendclient"
	"SkyeIM/app/message/rpc/messageclient"
	"SkyeIM/app/ws/internal/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	MessageRpc messageclient.Message
	FriendRpc  friendclient.Friend
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		MessageRpc: messageclient.NewMessage(zrpc.MustNewClient(c.MessageRpc)),
		FriendRpc:  friendclient.NewFriend(zrpc.MustNewClient(c.FriendRpc)),
	}
}
