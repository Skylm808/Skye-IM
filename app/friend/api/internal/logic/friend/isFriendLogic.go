// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package friend

import (
	"context"

	"SkyeIM/app/friend/api/internal/svc"
	"SkyeIM/app/friend/api/internal/types"
	"SkyeIM/app/friend/rpc/friendclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type IsFriendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 检查是否为好友
func NewIsFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IsFriendLogic {
	return &IsFriendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IsFriendLogic) IsFriend(req *types.GetFriendReq) (resp *types.IsFriendResp, err error) {
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.FriendRpc.IsFriend(l.ctx, &friendclient.IsFriendReq{
		UserId:   userId,
		FriendId: req.FriendId,
	})
	if err != nil {
		l.Logger.Errorf("IsFriend RPC error: %v", err)
		return nil, err
	}

	return &types.IsFriendResp{
		IsFriend: rpcResp.IsFriend,
		Status:   rpcResp.Status,
	}, nil
}
