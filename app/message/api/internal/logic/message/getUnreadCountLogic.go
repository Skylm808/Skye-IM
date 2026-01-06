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

type GetUnreadCountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取未读消息数
func NewGetUnreadCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnreadCountLogic {
	return &GetUnreadCountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUnreadCountLogic) GetUnreadCount(req *types.GetUnreadCountReq) (resp *types.GetUnreadCountResp, err error) {
	// 从 JWT 获取当前用户ID
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 调用 RPC 获取未读数
	rpcResp, err := l.svcCtx.MessageRpc.GetUnreadCount(l.ctx, &message.GetUnreadCountReq{
		UserId: userId,
		PeerId: req.PeerId,
	})
	if err != nil {
		l.Logger.Errorf("GetUnreadCount RPC failed: %v", err)
		return nil, err
	}

	return &types.GetUnreadCountResp{
		Count: rpcResp.Count,
	}, nil
}
