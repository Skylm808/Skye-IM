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
	list, err := l.svcCtx.ImMessageModel.GetUnreadMessages(l.ctx, in.UserId, in.PeerId)
	if err != nil {
		l.Logger.Errorf("GetUnreadMessages failed: %v", err)
		return nil, err
	}

	// 转换为响应格式
	respList := make([]*message.MessageInfo, 0, len(list))
	for _, msg := range list {
		respList = append(respList, &message.MessageInfo{
			Id:          msg.Id,
			MsgId:       msg.MsgId,
			FromUserId:  msg.FromUserId,
			ToUserId:    msg.ToUserId,
			Content:     msg.Content,
			ContentType: int32(msg.ContentType),
			Status:      int32(msg.Status),
			CreatedAt:   msg.CreatedAt.Unix(),
		})
	}

	return &message.GetUnreadMessagesResp{
		List: respList,
	}, nil
}
