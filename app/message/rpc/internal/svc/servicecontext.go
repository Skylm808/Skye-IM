package svc

import (
	"SkyeIM/app/message/rpc/internal/config"
	"SkyeIM/app/message/rpc/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config         config.Config
	ImMessageModel model.ImMessageModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)

	return &ServiceContext{
		Config:         c,
		ImMessageModel: model.NewImMessageModel(conn, c.Cache),
	}
}
