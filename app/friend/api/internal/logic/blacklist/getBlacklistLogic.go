// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package blacklist

import (
	"context"

	"SkyeIM/app/friend/api/internal/svc"
	"SkyeIM/app/friend/api/internal/types"
	"SkyeIM/app/friend/rpc/friendclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlacklistLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取黑名单列表
func NewGetBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlacklistLogic {
	return &GetBlacklistLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBlacklistLogic) GetBlacklist(req *types.PageReq) (resp *types.FriendListResp, err error) {
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.FriendRpc.GetBlacklist(l.ctx, &friendclient.GetBlacklistReq{
		UserId:   userId,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		l.Logger.Errorf("GetBlacklist RPC error: %v", err)
		return nil, err
	}

	list := make([]types.FriendInfo, 0, len(rpcResp.List))
	for _, item := range rpcResp.List {
		list = append(list, types.FriendInfo{
			Id:        item.Id,
			FriendId:  item.FriendId,
			Remark:    item.Remark,
			Status:    item.Status,
			CreatedAt: item.CreatedAt,
		})
	}

	return &types.FriendListResp{
		List:  list,
		Total: rpcResp.Total,
	}, nil
}
