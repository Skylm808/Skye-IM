// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package blacklist

import (
	"context"
	"encoding/json"
	"fmt"

	"SkyeIM/app/friend/api/internal/svc"
	"SkyeIM/app/friend/api/internal/types"
	"SkyeIM/app/friend/rpc/friendclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetBlacklistLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 设置黑名单
func NewSetBlacklistLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetBlacklistLogic {
	return &SetBlacklistLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetBlacklistLogic) SetBlacklist(req *types.SetBlacklistReq) (resp *types.Empty, err error) {
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.FriendRpc.SetBlacklist(l.ctx, &friendclient.SetBlacklistReq{
		UserId:   userId,
		FriendId: req.FriendId,
		IsBlack:  req.IsBlack,
	})
	if err != nil {
		l.Logger.Errorf("SetBlacklist RPC error: %v", err)
		return nil, err
	}

	return &types.Empty{}, nil
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
