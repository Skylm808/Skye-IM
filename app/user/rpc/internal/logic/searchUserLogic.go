package logic

import (
	"context"
	"database/sql"
	"strings"

	"SkyeIM/app/user/rpc/internal/svc"
	"SkyeIM/app/user/rpc/user"
	"auth/model"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SearchUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchUserLogic {
	return &SearchUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 搜索用户（用于添加好友）
func (l *SearchUserLogic) SearchUser(in *user.SearchUserRequest) (*user.SearchUserResponse, error) {
	keyword := strings.TrimSpace(in.Keyword)
	if keyword == "" {
		return nil, status.Error(codes.InvalidArgument, "搜索关键词不能为空")
	}

	var users []*user.UserInfo

	// 尝试按用户名精确匹配
	userInfo, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, keyword)
	if err == nil {
		users = append(users, convertToUserInfo(userInfo))
	} else if err != model.ErrNotFound {
		l.Logger.Errorf("按用户名搜索失败: %v", err)
	}

	// 尝试按邮箱精确匹配
	if strings.Contains(keyword, "@") {
		userInfo, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, sql.NullString{String: keyword, Valid: true})
		if err == nil {
			// 避免重复添加
			found := false
			for _, u := range users {
				if u.Id == int64(userInfo.Id) {
					found = true
					break
				}
			}
			if !found {
				users = append(users, convertToUserInfo(userInfo))
			}
		} else if err != model.ErrNotFound {
			l.Logger.Errorf("按邮箱搜索失败: %v", err)
		}
	}

	// 尝试按手机号精确匹配
	if len(keyword) >= 11 {
		userInfo, err := l.svcCtx.UserModel.FindOneByPhone(l.ctx, sql.NullString{String: keyword, Valid: true})
		if err == nil {
			found := false
			for _, u := range users {
				if u.Id == int64(userInfo.Id) {
					found = true
					break
				}
			}
			if !found {
				users = append(users, convertToUserInfo(userInfo))
			}
		} else if err != model.ErrNotFound {
			l.Logger.Errorf("按手机号搜索失败: %v", err)
		}
	}

	return &user.SearchUserResponse{
		Users: users,
		Total: int64(len(users)),
	}, nil
}
