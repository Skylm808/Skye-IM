package svc

import (
	groupModel "SkyeIM/app/group/model"
	"SkyeIM/app/user/api/internal/config"
	"SkyeIM/app/user/rpc/userClient"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	UserRpc    userClient.User
	GroupModel groupModel.ImGroupModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	// GroupModel暂时保留直接数据库访问（因为GlobalSearch需要）
	// TODO: 后续可以创建Group RPC服务
	conn := sqlx.NewMysql(c.MySQL.DataSource)
	return &ServiceContext{
		Config:     c,
		UserRpc:    userClient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		GroupModel: groupModel.NewImGroupModel(conn, c.Cache),
	}
}
