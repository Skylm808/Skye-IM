package profile

import (
	"SkyeIM/app/user/api/internal/types"
	"SkyeIM/app/user/rpc/userClient"
	"context"
	"encoding/json"
	"fmt"
)

// convertToUserInfo 转换RPC UserInfo为API types.UserInfo
func convertToUserInfo(u *userClient.UserInfo) types.UserInfo {
	return types.UserInfo{
		Id:        u.Id,
		Username:  u.Username,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Phone:     u.Phone,
		Email:     u.Email,
		Signature: u.Signature,
		Gender:    u.Gender,
		Region:    u.Region,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
	}
}

// getUserIdFromCtx 从上下文中获取用户 ID
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
