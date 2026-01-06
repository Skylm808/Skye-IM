// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"SkyeIM/app/message/api/internal/config"
	"SkyeIM/app/message/rpc/messageclient"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config     config.Config
	MessageRpc messageclient.Message
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		MessageRpc: messageclient.NewMessage(zrpc.MustNewClient(c.MessageRpc)),
	}
}
