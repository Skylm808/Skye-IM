// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package message

import (
	"context"

	"SkyeIM/app/message/api/internal/svc"
	"SkyeIM/app/message/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发送私聊消息（可选：主要走WS）
func NewSendMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendMessageLogic {
	return &SendMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendMessageLogic) SendMessage(req *types.SendMessageReq) (resp *types.SendMessageResp, err error) {
	// todo: add your logic here and delete this line

	return
}
