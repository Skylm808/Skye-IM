// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package message

import (
	"context"
	"encoding/json"
	"errors"

	"SkyeIM/app/message/api/internal/svc"
	"SkyeIM/app/message/api/internal/types"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMessageHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取与某用户的历史消息
func NewGetMessageHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMessageHistoryLogic {
	return &GetMessageHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMessageHistoryLogic) GetMessageHistory(req *types.GetMessageHistoryReq) (resp *types.GetMessageHistoryResp, err error) {
	// 从 JWT 获取当前用户ID
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 调用 RPC 获取消息列表
	rpcResp, err := l.svcCtx.MessageRpc.GetMessageList(l.ctx, &message.GetMessageListReq{
		UserId:    userId,
		PeerId:    req.PeerId,
		LastMsgId: req.LastMsgId,
		Limit:     req.Limit,
	})
	if err != nil {
		l.Logger.Errorf("GetMessageHistory RPC failed: %v", err)
		return nil, err
	}

	// 转换为 API 响应格式
	list := make([]types.MessageInfo, 0, len(rpcResp.List))
	for _, msg := range rpcResp.List {
		list = append(list, types.MessageInfo{
			Id:          msg.Id,
			MsgId:       msg.MsgId,
			FromUserId:  msg.FromUserId,
			ToUserId:    msg.ToUserId,
			ChatType:    msg.ChatType,
			GroupId:     msg.GroupId,
			Content:     msg.Content,
			ContentType: msg.ContentType,
			Status:      msg.Status,
			CreatedAt:   msg.CreatedAt,
			Seq:         msg.Seq,
		})
	}

	return &types.GetMessageHistoryResp{
		List:    list,
		HasMore: rpcResp.HasMore,
	}, nil
}

// getUserIdFromCtx 从 context 获取用户ID
func getUserIdFromCtx(ctx context.Context) (int64, error) {
	userId := ctx.Value("userId")
	if userId == nil {
		return 0, errors.New("userId not found in context")
	}
	switch v := userId.(type) {
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case json.Number:
		return v.Int64()
	default:
		return 0, errors.New("invalid userId type")
	}
}
