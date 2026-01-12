package logic

import (
	"context"

	"SkyeIM/app/message/rpc/internal/svc"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMessageListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMessageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessageListLogic {
	return &GetMessageListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取私聊历史消息列表
func (l *GetMessageListLogic) GetMessageList(in *message.GetMessageListReq) (*message.GetMessageListResp, error) {
	limit := in.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// 调用Model层方法
	messages, err := l.svcCtx.ImMessageModel.FindPrivateMessageList(
		l.ctx,
		in.UserId,
		in.PeerId,
		in.LastMsgId,
		int64(limit)+1,
	)
	if err != nil {
		l.Logger.Errorf("查询私聊消息失败: %v", err)
		return nil, err
	}

	hasMore := false
	if int64(len(messages)) > int64(limit) {
		hasMore = true
		messages = messages[:limit]
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

	return &message.GetMessageListResp{
		List:    list,
		HasMore: hasMore,
	}, nil
}
