// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package request

import (
	"context"
	"encoding/json"
	"fmt"

	"SkyeIM/app/friend/api/internal/svc"
	"SkyeIM/app/friend/api/internal/types"
	"SkyeIM/app/friend/rpc/friendclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddFriendRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发送好友申请
func NewAddFriendRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddFriendRequestLogic {
	return &AddFriendRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddFriendRequestLogic) AddFriendRequest(req *types.AddFriendRequestReq) (resp *types.AddFriendRequestResp, err error) {
	// 从JWT获取当前用户ID
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 调用RPC服务
	rpcResp, err := l.svcCtx.FriendRpc.AddFriendRequest(l.ctx, &friendclient.AddFriendRequestReq{
		FromUserId: userId,
		ToUserId:   req.ToUserId,
		Message:    req.Message,
	})
	if err != nil {
		l.Logger.Errorf("AddFriendRequest RPC error: %v", err)
		return nil, err
	}

	return &types.AddFriendRequestResp{
		RequestId: rpcResp.RequestId,
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
