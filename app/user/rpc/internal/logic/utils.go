package logic

import (
	"SkyeIM/app/user/rpc/user"
	"auth/model"
)

// convertToUserInfo 将 model.User 转换为 proto 的 UserInfo
func convertToUserInfo(u *model.User) *user.UserInfo {
	phone := ""
	if u.Phone.Valid {
		phone = u.Phone.String
	}
	email := ""
	if u.Email.Valid {
		email = u.Email.String
	}

	return &user.UserInfo{
		Id:        int64(u.Id),
		Username:  u.Username,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Phone:     phone,
		Email:     email,
		Signature: u.Signature,
		Gender:    int64(u.Gender),
		Region:    u.Region,
		Status:    int64(u.Status),
		CreatedAt: u.CreatedAt.Unix(),
	}
}
