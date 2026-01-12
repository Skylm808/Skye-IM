package logic

import (
	"context"
	"encoding/json"

	"SkyeIM/app/message/rpc/internal/svc"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchMessageLogic {
	return &SearchMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 模糊搜索消息内容
func (l *SearchMessageLogic) SearchMessage(in *message.SearchMessageReq) (*message.SearchMessageResp, error) {
	messages, err := l.svcCtx.ImMessageModel.SearchByKeyword(l.ctx, in.UserId, in.Keyword)
	if err != nil {
		l.Logger.Errorf("RPC 搜索消息失败: %v", err)
		return nil, err
	}

	var list []*message.MessageInfo
	for _, v := range messages {
		// 解析at_user_ids JSON
		var atUserIds []int64
		if v.AtUserIds.Valid && v.AtUserIds.String != "" {
			_ = json.Unmarshal([]byte(v.AtUserIds.String), &atUserIds)
		}

		list = append(list, &message.MessageInfo{
			Id:          int64(v.Id),
			MsgId:       v.MsgId,
			FromUserId:  int64(v.FromUserId),
			ToUserId:    int64(v.ToUserId),
			ChatType:    int32(v.ChatType),
			GroupId:     v.GroupId.String,
			Content:     v.Content,
			ContentType: int32(v.ContentType),
			Status:      int32(v.Status),
			CreatedAt:   v.CreatedAt.Unix(),
			Seq:         v.Seq,
			AtUserIds:   atUserIds,
		})
	}

	return &message.SearchMessageResp{
		List: list,
	}, nil
}
