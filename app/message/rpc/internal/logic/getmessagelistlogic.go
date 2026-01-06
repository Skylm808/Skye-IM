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

// 获取历史消息列表（分页）
func (l *GetMessageListLogic) GetMessageList(in *message.GetMessageListReq) (*message.GetMessageListResp, error) {
	// 默认每页20条
	limit := int(in.Limit)
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 查询消息列表（多取一条用于判断是否有更多）
	list, err := l.svcCtx.ImMessageModel.GetMessageList(l.ctx, in.UserId, in.PeerId, in.LastMsgId, limit+1)
	if err != nil {
		l.Logger.Errorf("GetMessageList failed: %v", err)
		return nil, err
	}

	// 判断是否有更多
	hasMore := len(list) > limit
	if hasMore {
		list = list[:limit]
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

	return &message.GetMessageListResp{
		List:    respList,
		HasMore: hasMore,
	}, nil
}
