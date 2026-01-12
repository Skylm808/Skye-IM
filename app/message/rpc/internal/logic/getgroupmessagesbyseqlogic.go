package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	// 1. 查询消息
	messages, err := l.svcCtx.ImMessageModel.FindGroupMessagesAfterSeq(l.ctx, in.GroupId, in.Seq)
	if err != nil {
		l.Logger.Errorf("同步群消息失败: %v", err)
		return nil, nil // Return empty or error
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
			// 确保 seq 字段在 proto 中也加上了
			Seq: msg.Seq,
		})
	}

	return &message.GetGroupMessagesBySeqResp{
		List: list,
	}, nil
}
