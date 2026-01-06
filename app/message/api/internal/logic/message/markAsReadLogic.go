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

type MarkAsReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 标记消息为已读
func NewMarkAsReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkAsReadLogic {
	return &MarkAsReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkAsReadLogic) MarkAsRead(req *types.MarkAsReadReq) (resp *types.MarkAsReadResp, err error) {
	// 从 JWT 获取当前用户ID
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 调用 RPC 标记已读
	rpcResp, err := l.svcCtx.MessageRpc.MarkAsRead(l.ctx, &message.MarkAsReadReq{
		UserId: userId,
		PeerId: req.PeerId,
		MsgIds: req.MsgIds,
	})
	if err != nil {
		l.Logger.Errorf("MarkAsRead RPC failed: %v", err)
		return nil, err
	}

	return &types.MarkAsReadResp{
		Count: rpcResp.Count,
	}, nil
}
