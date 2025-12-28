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

	return &types.ProfileResponse{
		User: convertToUserInfo(user),
	}, nil
}

// getUserIdFromCtx 从上下文中获取用户 ID
func (l *GetProfileLogic) getUserIdFromCtx() (int64, error) {
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

// convertToUserInfo 转换为响应类型
func convertToUserInfo(u *model.User) types.UserInfo {
	phone := ""
	if u.Phone.Valid {
		phone = u.Phone.String
	}
	email := u.Email.String

	return types.UserInfo{
		Id:        int64(u.Id),
		Username:  u.Username,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Phone:     phone,
		Email:     email,
		Status:    int64(u.Status),
		CreatedAt: u.CreatedAt.Unix(),
	}
}
