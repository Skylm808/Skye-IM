package logic

import (
	"context"

	"SkyeIM/app/message/rpc/internal/svc"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkAsReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkAsReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkAsReadLogic {
	return &MarkAsReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 标记私聊消息为已读
func (l *MarkAsReadLogic) MarkAsRead(in *message.MarkAsReadReq) (*message.MarkAsReadResp, error) {
	// 调用Model层方法
	count, err := l.svcCtx.ImMessageModel.MarkMessagesAsRead(
		l.ctx,
		in.UserId,
		in.PeerId,
		in.MsgIds,
	)
	if err != nil {
		l.Logger.Errorf("标记已读失败: %v", err)
		return nil, err
	}

	return &message.MarkAsReadResp{
		Count: count,
	}, nil
}
