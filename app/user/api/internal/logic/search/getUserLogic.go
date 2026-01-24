// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package search

import (
	"context"
	"fmt"

	"SkyeIM/app/user/api/internal/svc"
	"SkyeIM/app/user/api/internal/types"
	"SkyeIM/app/user/rpc/userClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取指定用户信息
func NewGetUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLogic {
	return &GetUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserLogic) GetUser(req *types.GetUserRequest) (resp *types.ProfileResponse, err error) {
	if req.Id <= 0 {
		return nil, fmt.Errorf("无效的用户ID")
	}

	// 通过RPC获取用户信息
	userResp, err := l.svcCtx.UserRpc.GetUser(l.ctx, &userClient.GetUserRequest{
		Id: req.Id,
	})
	if err != nil {
		l.Logger.Errorf("RPC获取用户失败: %v", err)
		return nil, fmt.Errorf("获取用户信息失败")
	}

	return &types.ProfileResponse{
		User: convertToUserInfo(userResp.User),
	}, nil
}
