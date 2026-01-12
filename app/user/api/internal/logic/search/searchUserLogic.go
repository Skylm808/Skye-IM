// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package search

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"SkyeIM/app/user/api/internal/svc"
	"SkyeIM/app/user/api/internal/types"
	"auth/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 搜索用户（用于添加好友）
func NewSearchUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchUserLogic {
	return &SearchUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchUserLogic) SearchUser(req *types.SearchUserRequest) (resp *types.SearchUserResponse, err error) {
	keyword := strings.TrimSpace(req.Keyword)
	if keyword == "" {
		return nil, fmt.Errorf("搜索关键词不能为空")
	}

	var users []types.UserInfo

	// 尝试按用户名精确匹配
	user, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, keyword)
	if err == nil {
		users = append(users, convertToUserInfo(user))
	} else if err != model.ErrNotFound {
		l.Logger.Errorf("按用户名搜索失败: %v", err)
	}

	// 尝试按邮箱精确匹配
	if strings.Contains(keyword, "@") {
		user, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, sql.NullString{String: keyword, Valid: true})
		if err == nil {
			// 避免重复添加
			found := false
			for _, u := range users {
				if u.Id == int64(user.Id) {
					found = true
					break
				}
			}
			if !found {
				users = append(users, convertToUserInfo(user))
			}
		} else if err != model.ErrNotFound {
			l.Logger.Errorf("按邮箱搜索失败: %v", err)
		}
	}

	// 尝试按手机号精确匹配
	if len(keyword) >= 11 {
		user, err := l.svcCtx.UserModel.FindOneByPhone(l.ctx, sql.NullString{String: keyword, Valid: true})
		if err == nil {
			found := false
			for _, u := range users {
				if u.Id == int64(user.Id) {
					found = true
					break
				}
			}
			if !found {
				users = append(users, convertToUserInfo(user))
			}
		} else if err != model.ErrNotFound {
			l.Logger.Errorf("按手机号搜索失败: %v", err)
		}
	}

	return &types.SearchUserResponse{
		Users: users,
		Total: int64(len(users)),
	}, nil
}

// convertToUserInfo 转换为响应类型
func convertToUserInfo(u *model.User) types.UserInfo {
	phone := ""
	if u.Phone.Valid {
		phone = u.Phone.String
	}
	email := ""
	if u.Email.Valid {
		email = u.Email.String
	}

	return types.UserInfo{
		Id:        int64(u.Id),
		Username:  u.Username,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Signature: u.Signature,
		Gender:    int64(u.Gender),
		Region:    u.Region,
		Phone:     phone,
		Email:     email,
		Status:    int64(u.Status),
		CreatedAt: u.CreatedAt.Unix(),
	}
}
