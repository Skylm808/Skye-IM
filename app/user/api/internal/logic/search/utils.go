package search

import (
	"SkyeIM/app/user/api/internal/types"
	"SkyeIM/app/user/rpc/userClient"
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
