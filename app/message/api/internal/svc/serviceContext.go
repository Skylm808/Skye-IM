// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"SkyeIM/app/friend/rpc/friendclient"
	"SkyeIM/app/group/rpc/groupclient"
	"SkyeIM/app/message/api/internal/config"
	"SkyeIM/app/message/rpc/messageclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	MessageRpc messageclient.Message
	GroupRpc   groupclient.Group
	FriendRpc  friendclient.Friend
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		MessageRpc: messageclient.NewMessage(zrpc.MustNewClient(c.MessageRpc)),
		GroupRpc:   groupclient.NewGroup(zrpc.MustNewClient(c.GroupRpc)),
		FriendRpc:  friendclient.NewFriend(zrpc.MustNewClient(c.FriendRpc)),
	}
}
