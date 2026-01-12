package svc

import (
	groupModel "SkyeIM/app/group/model"
	"SkyeIM/app/user/api/internal/config"
	"auth/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config     config.Config
	UserModel  model.UserModel
	GroupModel groupModel.ImGroupModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)
	return &ServiceContext{
		Config:     c,
		UserModel:  model.NewUserModel(conn, c.Cache),
		GroupModel: groupModel.NewImGroupModel(conn, c.Cache),
	}
}
