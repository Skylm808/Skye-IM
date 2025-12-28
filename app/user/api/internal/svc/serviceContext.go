// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"SkyeIM/app/user/api/internal/config"
	"auth/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config
	UserModel model.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)
	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(conn, c.Cache),
	}
}
