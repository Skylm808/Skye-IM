// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package profile

import (
	"context"
	"encoding/json"
	"fmt"

	"SkyeIM/app/user/api/internal/svc"
	"SkyeIM/app/user/api/internal/types"
	"auth/model"

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
	userId, err := l.getUserIdFromCtx()
	if err != nil {
		return nil, err
	}

	// 查询用户信息
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(userId))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, fmt.Errorf("用户不存在")
		}
		l.Logger.Errorf("查询用户失败: %v", err)
		return nil, fmt.Errorf("查询用户失败")
	}

	// 更新头像
	user.Avatar = req.Avatar

	// 保存更新
	err = l.svcCtx.UserModel.Update(l.ctx, user)
	if err != nil {
		l.Logger.Errorf("更新头像失败: %v", err)
		return nil, fmt.Errorf("更新头像失败")
	}

	return &types.ProfileResponse{
		User: convertToUserInfo(user),
	}, nil
}

// getUserIdFromCtx 从上下文中获取用户 ID
func (l *UpdateAvatarLogic) getUserIdFromCtx() (int64, error) {
	userId := l.ctx.Value("userId")
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
