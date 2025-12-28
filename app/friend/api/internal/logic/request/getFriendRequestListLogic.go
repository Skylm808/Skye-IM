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

type GetFriendRequestListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取收到的好友申请列表
func NewGetFriendRequestListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendRequestListLogic {
	return &GetFriendRequestListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFriendRequestListLogic) GetFriendRequestList(req *types.PageReq) (resp *types.FriendRequestListResp, err error) {
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.FriendRpc.GetFriendRequestList(l.ctx, &friendclient.GetFriendRequestListReq{
		UserId:   userId,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		l.Logger.Errorf("GetFriendRequestList RPC error: %v", err)
		return nil, err
	}

	list := make([]types.FriendRequestInfo, 0, len(rpcResp.List))
	for _, item := range rpcResp.List {
		list = append(list, types.FriendRequestInfo{
			Id:         item.Id,
			FromUserId: item.FromUserId,
			ToUserId:   item.ToUserId,
			Message:    item.Message,
			Status:     item.Status,
			CreatedAt:  item.CreatedAt,
		})
	}

	return &types.FriendRequestListResp{
		List:  list,
		Total: rpcResp.Total,
	}, nil
}
