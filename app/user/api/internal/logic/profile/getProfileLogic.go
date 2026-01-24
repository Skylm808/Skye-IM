// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package profile

import (
	"context"
	"fmt"

	"SkyeIM/app/user/api/internal/svc"
	"SkyeIM/app/user/api/internal/types"
	"SkyeIM/app/user/rpc/userClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取当前用户资料
func NewGetProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProfileLogic {
	return &GetProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProfileLogic) GetProfile() (resp *types.ProfileResponse, err error) {
	// 从 JWT 中获取用户 ID
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 通过RPC获取用户信息
	userResp, err := l.svcCtx.UserRpc.GetUser(l.ctx, &userClient.GetUserRequest{
		Id: userId,
	})
	if err != nil {
		l.Logger.Errorf("RPC获取用户失败: %v", err)
		return nil, fmt.Errorf("获取用户信息失败")
	}

	return &types.ProfileResponse{
		User: convertToUserInfo(userResp.User),
	}, nil
}
