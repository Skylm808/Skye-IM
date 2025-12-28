// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package friend

import (
	"context"
	"encoding/json"
	"fmt"

	"SkyeIM/app/friend/api/internal/svc"
	"SkyeIM/app/friend/api/internal/types"
	"SkyeIM/app/friend/rpc/friendclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFriendListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取好友列表
func NewGetFriendListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFriendListLogic {
	return &GetFriendListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFriendListLogic) GetFriendList(req *types.PageReq) (resp *types.FriendListResp, err error) {
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	rpcResp, err := l.svcCtx.FriendRpc.GetFriendList(l.ctx, &friendclient.GetFriendListReq{
		UserId:   userId,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		l.Logger.Errorf("GetFriendList RPC error: %v", err)
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

// getUserIdFromCtx 从上下文中获取用户ID
func getUserIdFromCtx(ctx context.Context) (int64, error) {
	userId := ctx.Value("userId")
	if userId == nil {
		return 0, fmt.Errorf("未登录")
	}

	switch v := userId.(type) {
	case json.Number:
		return v.Int64()
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	default:
		return 0, fmt.Errorf("无效的用户ID类型")
	}
}
