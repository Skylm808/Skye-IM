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

type UpdateAvatarLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新头像
func NewUpdateAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAvatarLogic {
	return &UpdateAvatarLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateAvatarLogic) UpdateAvatar(req *types.UpdateAvatarRequest) (resp *types.ProfileResponse, err error) {
	// 从 JWT 中获取用户 ID
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 通过RPC更新头像
	userResp, err := l.svcCtx.UserRpc.UpdateUser(l.ctx, &userClient.UpdateUserRequest{
		Id:     userId,
		Avatar: req.Avatar,
	})
	if err != nil {
		l.Logger.Errorf("RPC更新头像失败: %v", err)
		return nil, fmt.Errorf("更新头像失败")
	}

	return &types.ProfileResponse{
		User: convertToUserInfo(userResp.User),
	}, nil
}
