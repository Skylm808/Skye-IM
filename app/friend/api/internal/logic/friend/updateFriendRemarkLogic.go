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

type UpdateFriendRemarkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新好友备注
func NewUpdateFriendRemarkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateFriendRemarkLogic {
	return &UpdateFriendRemarkLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateFriendRemarkLogic) UpdateFriendRemark(req *types.UpdateRemarkReq) (resp *types.Empty, err error) {
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.FriendRpc.UpdateFriendRemark(l.ctx, &friendclient.UpdateFriendRemarkReq{
		UserId:   userId,
		FriendId: req.FriendId,
		Remark:   req.Remark,
	})
	if err != nil {
		l.Logger.Errorf("UpdateFriendRemark RPC error: %v", err)
		return nil, err
	}

	return &types.Empty{}, nil
}
