// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package message

import (
	"context"

	"SkyeIM/app/message/api/internal/svc"
	"SkyeIM/app/message/api/internal/types"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupMessageHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取群聊历史消息
func NewGetGroupMessageHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupMessageHistoryLogic {
	return &GetGroupMessageHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGroupMessageHistoryLogic) GetGroupMessageHistory(req *types.GetGroupMessageHistoryReq) (resp *types.GetGroupMessageHistoryResp, err error) {
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 调用 RPC 获取群消息列表
	rpcLimit := req.Limit
	if rpcLimit <= 0 {
		rpcLimit = 20
	}

	rpcResp, err := l.svcCtx.MessageRpc.GetGroupMessageList(l.ctx, &message.GetGroupMessageListReq{
		UserId:    userId,
		GroupId:   req.GroupId,
		LastMsgId: req.LastMsgId,
		Limit:     rpcLimit,
	})

	if err != nil {
		l.Logger.Errorf("获取群历史消息失败: %v", err)
		return nil, err
	}

	list := make([]types.MessageInfo, 0, len(rpcResp.List))
	for _, msg := range rpcResp.List {
		list = append(list, types.MessageInfo{
			Id:          msg.Id,
			MsgId:       msg.MsgId,
			FromUserId:  msg.FromUserId,
			ToUserId:    msg.ToUserId, // 群聊通常为0
			ChatType:    msg.ChatType,
			GroupId:     msg.GroupId,
			Content:     msg.Content,
			ContentType: msg.ContentType,
			Status:      msg.Status,
			CreatedAt:   msg.CreatedAt,
			Seq:         msg.Seq,
		})
	}

	return &types.GetGroupMessageHistoryResp{
		List:    list,
		HasMore: rpcResp.HasMore,
	}, nil
}
