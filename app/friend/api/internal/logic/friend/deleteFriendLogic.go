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

type DeleteFriendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除好友
func NewDeleteFriendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteFriendLogic {
	return &DeleteFriendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteFriendLogic) DeleteFriend(req *types.DeleteFriendReq) (resp *types.Empty, err error) {
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.FriendRpc.DeleteFriend(l.ctx, &friendclient.DeleteFriendReq{
		UserId:   userId,
		FriendId: req.FriendId,
	})
	if err != nil {
		l.Logger.Errorf("DeleteFriend RPC error: %v", err)
		return nil, err
	}

	return &types.Empty{}, nil
}
