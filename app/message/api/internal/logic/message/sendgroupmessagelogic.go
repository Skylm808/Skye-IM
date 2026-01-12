// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package message

import (
	"context"

	"SkyeIM/app/message/api/internal/svc"
	"SkyeIM/app/message/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendGroupMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发送群聊消息（可选：主要走WS）
func NewSendGroupMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendGroupMessageLogic {
	return &SendGroupMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendGroupMessageLogic) SendGroupMessage(req *types.SendGroupMessageReq) (resp *types.SendGroupMessageResp, err error) {
	// todo: add your logic here and delete this line

	return
}
