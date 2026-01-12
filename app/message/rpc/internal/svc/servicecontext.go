package svc

import (
	"SkyeIM/app/group/rpc/groupclient"
	"SkyeIM/app/message/model"
	"SkyeIM/app/message/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	ImMessageModel model.ImMessageModel
	GroupRpc       groupclient.Group
	Redis          *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)

	return &ServiceContext{
		Config:         c,
		ImMessageModel: model.NewImMessageModel(conn, c.Cache),
		GroupRpc:       groupclient.NewGroup(zrpc.MustNewClient(c.GroupRpc)),
		Redis:          redis.MustNewRedis(c.Cache[0].RedisConf),
	}
}
