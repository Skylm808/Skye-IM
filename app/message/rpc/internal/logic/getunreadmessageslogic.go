package logic

import (
	"context"

	"SkyeIM/app/message/rpc/internal/svc"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnreadMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUnreadMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadMessagesLogic {
	return &GetUnreadMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取与某用户的未读消息
func (l *GetUnreadMessagesLogic) GetUnreadMessages(in *message.GetUnreadMessagesReq) (*message.GetUnreadMessagesResp, error) {
	// 调用Model层方法
	messages, err := l.svcCtx.ImMessageModel.FindUnreadMessages(
		l.ctx,
		in.UserId,
		in.PeerId,
	)
	if err != nil {
		l.Logger.Errorf("查询未读消息失败: %v", err)
		return nil, err
	}

	var list []*message.MessageInfo
	for _, msg := range messages {
		list = append(list, &message.MessageInfo{
			Id:          int64(msg.Id),
			MsgId:       msg.MsgId,
			FromUserId:  int64(msg.FromUserId),
			ToUserId:    int64(msg.ToUserId),
			ChatType:    int32(msg.ChatType),
			GroupId:     msg.GroupId.String,
			Content:     msg.Content,
			ContentType: int32(msg.ContentType),
			Status:      int32(msg.Status),
			CreatedAt:   msg.CreatedAt.Unix(),
		})
	}

	return &message.GetUnreadMessagesResp{
		List: list,
	}, nil
}
