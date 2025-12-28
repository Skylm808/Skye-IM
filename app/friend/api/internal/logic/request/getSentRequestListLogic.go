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

type GetSentRequestListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取发出的好友申请列表
func NewGetSentRequestListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSentRequestListLogic {
	return &GetSentRequestListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSentRequestListLogic) GetSentRequestList(req *types.PageReq) (resp *types.FriendRequestListResp, err error) {
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.FriendRpc.GetSentRequestList(l.ctx, &friendclient.GetSentRequestListReq{
		UserId:   userId,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		l.Logger.Errorf("GetSentRequestList RPC error: %v", err)
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
