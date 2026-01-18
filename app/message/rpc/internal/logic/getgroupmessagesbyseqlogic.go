package logic

import (
	"context"
	"encoding/json"

	"SkyeIM/app/message/rpc/internal/svc"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupMessagesBySeqLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupMessagesBySeqLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupMessagesBySeqLogic {
	return &GetGroupMessagesBySeqLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取大于指定Seq的群聊消息 (用于消息同步)
func (l *GetGroupMessagesBySeqLogic) GetGroupMessagesBySeq(in *message.GetGroupMessagesBySeqReq) (*message.GetGroupMessagesBySeqResp, error) {
	// 1. 查询消息
	messages, err := l.svcCtx.ImMessageModel.FindGroupMessagesAfterSeq(l.ctx, in.GroupId, in.Seq)
	if err != nil {
		l.Logger.Errorf("同步群消息失败: %v", err)
		return nil, err // 修复：返回错误而不是 nil
	}

	var list []*message.MessageInfo
	for _, msg := range messages {
		// 解析 @用户列表
		var atUserIds []int64
		if msg.AtUserIds.Valid && msg.AtUserIds.String != "" {
			// 从 JSON 解析到 []int64
			var ids []int64
			if err := json.Unmarshal([]byte(msg.AtUserIds.String), &ids); err == nil {
				atUserIds = ids
			} else {
				l.Logger.Errorf("解析 AtUserIds 失败，msg_id=%s: %v", msg.MsgId, err)
			}
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
			AtUserIds:   atUserIds, // 修复：添加 AtUserIds 字段
		})
	}

	return &message.GetGroupMessagesBySeqResp{
		List: list,
	}, nil
}
