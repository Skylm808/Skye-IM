package svc

import (
	"SkyeIM/app/group/model"
	"SkyeIM/app/group/rpc/internal/config"
	"SkyeIM/common/wspush"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                  config.Config
	Redis                   *redis.Redis
	ImGroupModel            model.ImGroupModel
	ImGroupMemberModel      model.ImGroupMemberModel
	ImGroupInvitationModel  model.ImGroupInvitationModel
	ImGroupJoinRequestModel model.ImGroupJoinRequestModel
	WsPushClient            *wspush.WsPushClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)
	return &ServiceContext{
		Config: c,
		Redis: redis.MustNewRedis(redis.RedisConf{
			Host: c.Cache[0].Host,
			Type: c.Cache[0].Type,
			Pass: c.Cache[0].Pass,
		}),
		ImGroupModel:            model.NewImGroupModel(conn, c.Cache),
		ImGroupMemberModel:      model.NewImGroupMemberModel(conn, c.Cache),
		ImGroupInvitationModel:  model.NewImGroupInvitationModel(conn, c.Cache),
		ImGroupJoinRequestModel: model.NewImGroupJoinRequestModel(conn, c.Cache),
		WsPushClient:            wspush.NewWsPushClient(c.WsServiceUrl, c.WsPushSecret),
	}
}
