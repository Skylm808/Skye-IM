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

type UpdateProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新用户资料
func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProfileLogic) UpdateProfile(req *types.UpdateProfileRequest) (resp *types.ProfileResponse, err error) {
	// 从 JWT 中获取用户 ID
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 通过RPC更新用户信息
	userResp, err := l.svcCtx.UserRpc.UpdateUser(l.ctx, &userClient.UpdateUserRequest{
		Id:        userId,
		Nickname:  req.Nickname,
		Avatar:    req.Avatar,
		Phone:     req.Phone,
		Signature: req.Signature,
		Gender:    req.Gender,
		Region:    req.Region,
	})
	if err != nil {
		l.Logger.Errorf("RPC更新用户失败: %v", err)
		return nil, fmt.Errorf("更新用户资料失败")
	}

	return &types.ProfileResponse{
		User: convertToUserInfo(userResp.User),
	}, nil
}
