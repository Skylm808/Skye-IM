package logic

import (
	"context"

	"SkyeIM/app/message/rpc/internal/svc"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnreadCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUnreadCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadCountLogic {
	return &GetUnreadCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取私聊未读消息数量
func (l *GetUnreadCountLogic) GetUnreadCount(in *message.GetUnreadCountReq) (*message.GetUnreadCountResp, error) {
	// 调用Model层方法
	count, err := l.svcCtx.ImMessageModel.CountUnreadMessages(
		l.ctx,
		in.UserId,
		in.PeerId,
	)
	if err != nil {
		l.Logger.Errorf("查询未读消息数失败: %v", err)
		return nil, err
	}

	return &message.GetUnreadCountResp{
		Count: count,
	}, nil
}
