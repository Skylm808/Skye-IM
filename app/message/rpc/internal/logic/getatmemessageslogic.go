package logic

import (
	"context"
	"encoding/json"

	"SkyeIM/app/message/rpc/internal/svc"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetAtMeMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAtMeMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAtMeMessagesLogic {
	return &GetAtMeMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取@我的消息列表
func (l *GetAtMeMessagesLogic) GetAtMeMessages(in *message.GetAtMeMessagesReq) (*message.GetAtMeMessagesResp, error) {
	if in.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "用户ID不能为空")
	}

	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}

	// 调用Model层查询@我的消息
	messages, err := l.svcCtx.ImMessageModel.FindAtMeMessages(
		l.ctx,
		in.UserId,
		in.GroupId,
		in.LastMsgId,
		limit,
	)
	if err != nil {
		l.Logger.Errorf("查询@我的消息失败: %v", err)
		return nil, status.Error(codes.Internal, "查询失败")
	}

	// 转换为proto格式
	list := make([]*message.MessageInfo, 0, len(messages))
	for _, msg := range messages {
		// 解析at_user_ids JSON
		var atUserIds []int64
		if msg.AtUserIds.Valid && msg.AtUserIds.String != "" {
			_ = json.Unmarshal([]byte(msg.AtUserIds.String), &atUserIds)
		}

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
			Seq:         msg.Seq,
			AtUserIds:   atUserIds,
		})
	}

	return &message.GetAtMeMessagesResp{
		List:    list,
		HasMore: len(messages) >= int(limit),
	}, nil
}
