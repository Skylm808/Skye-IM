// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package message

import (
	"context"
	"encoding/json"

	"SkyeIM/app/message/api/internal/svc"
	"SkyeIM/app/message/api/internal/types"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAtMeMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取@我的消息列表
func NewGetAtMeMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAtMeMessagesLogic {
	return &GetAtMeMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAtMeMessagesLogic) GetAtMeMessages(req *types.GetAtMeMessagesReq) (resp *types.GetAtMeMessagesResp, err error) {
	// 从JWT获取用户ID
	userId, err := l.ctx.Value("userId").(json.Number).Int64()
	if err != nil {
		return nil, err
	}

	// 调用Message RPC
	rpcResp, err := l.svcCtx.MessageRpc.GetAtMeMessages(l.ctx, &message.GetAtMeMessagesReq{
		UserId:    userId,
		GroupId:   req.GroupId,
		LastMsgId: req.LastMsgId,
		Limit:     req.Limit,
	})
	if err != nil {
		l.Logger.Errorf("查询@我的消息失败: %v", err)
		return nil, err
	}

	// 转换为API格式
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
			AtUserIds:   msg.AtUserIds,
		})
	}

	return &types.GetAtMeMessagesResp{
		List:    list,
		HasMore: rpcResp.HasMore,
	}, nil
}
