package svc

import (
	"SkyeIM/app/friend/rpc/friendclient"
	"SkyeIM/app/group/rpc/groupclient"
	"SkyeIM/app/message/rpc/messageclient"
	"SkyeIM/app/ws/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	Redis      *redis.Redis
	MessageRpc messageclient.Message
	FriendRpc  friendclient.Friend
	GroupRpc   groupclient.Group
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		Redis:      redis.MustNewRedis(redis.RedisConf{Host: c.Redis.Host, Type: c.Redis.Type, Pass: c.Redis.Pass}),
		MessageRpc: messageclient.NewMessage(zrpc.MustNewClient(c.MessageRpc)),
		FriendRpc:  friendclient.NewFriend(zrpc.MustNewClient(c.FriendRpc)),
		GroupRpc:   groupclient.NewGroup(zrpc.MustNewClient(c.GroupRpc)),
	}
}
