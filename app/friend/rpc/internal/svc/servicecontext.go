package svc

import (
	"SkyeIM/app/friend/rpc/internal/config"
	"SkyeIM/app/friend/rpc/model"
	"SkyeIM/common/wspush"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config             config.Config
	FriendModel        model.ImFriendModel
	FriendRequestModel model.ImFriendRequestModel
	WsPushClient       *wspush.WsPushClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)
	return &ServiceContext{
		Config:             c,
		FriendModel:        model.NewImFriendModel(conn, c.Cache),
		FriendRequestModel: model.NewImFriendRequestModel(conn, c.Cache),
		WsPushClient:       wspush.NewWsPushClient(c.WsServiceUrl, c.WsPushSecret),
	}
}
