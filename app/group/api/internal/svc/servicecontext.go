// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"SkyeIM/app/group/api/internal/config"
	"SkyeIM/app/group/rpc/groupclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config   config.Config
	GroupRpc groupclient.Group
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:   c,
		GroupRpc: groupclient.NewGroup(zrpc.MustNewClient(c.GroupRpc)),
	}
}
