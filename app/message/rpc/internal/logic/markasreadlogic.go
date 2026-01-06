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

// 标记消息为已读
func (l *MarkAsReadLogic) MarkAsRead(in *message.MarkAsReadReq) (*message.MarkAsReadResp, error) {
	count, err := l.svcCtx.ImMessageModel.MarkAsRead(l.ctx, in.UserId, in.PeerId, in.MsgIds)
	if err != nil {
		l.Logger.Errorf("MarkAsRead failed: %v", err)
		return nil, err
	}

	return &message.MarkAsReadResp{
		Count: count,
	}, nil
}
