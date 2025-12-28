// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package request

import (
	"context"

	"SkyeIM/app/friend/api/internal/svc"
	"SkyeIM/app/friend/api/internal/types"
	"SkyeIM/app/friend/rpc/friendclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleFriendRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 处理好友申请
func NewHandleFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleFriendRequestLogic {
	return &HandleFriendRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleFriendRequestLogic) HandleFriendRequest(req *types.HandleFriendRequestReq) (resp *types.Empty, err error) {
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.FriendRpc.HandleFriendRequest(l.ctx, &friendclient.HandleFriendRequestReq{
		UserId:    userId,
		RequestId: req.RequestId,
		Action:    req.Action,
	})
	if err != nil {
		l.Logger.Errorf("HandleFriendRequest RPC error: %v", err)
		return nil, err
	}

	return &types.Empty{}, nil
}
