package logic

import (
	"context"

	"SkyeIM/app/user/rpc/internal/svc"
	"SkyeIM/app/user/rpc/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchUsersByKeywordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchUsersByKeywordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchUsersByKeywordLogic {
	return &SearchUsersByKeywordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 模糊搜索用户（全局搜索用）
func (l *SearchUsersByKeywordLogic) SearchUsersByKeyword(in *user.SearchUsersByKeywordRequest) (*user.SearchUsersByKeywordResponse, error) {
	// 模糊搜索用户
	users, err := l.svcCtx.UserModel.SearchByKeyword(l.ctx, in.Keyword)
	if err != nil {
		l.Logger.Errorf("模糊搜索用户失败: keyword=%s, error=%v", in.Keyword, err)
		return &user.SearchUsersByKeywordResponse{
			Users: []*user.UserInfo{},
		}, nil
	}

	// 转换结果
	var userList []*user.UserInfo
	for _, u := range users {
		userList = append(userList, convertToUserInfo(u))
	}

	return &user.SearchUsersByKeywordResponse{
		Users: userList,
	}, nil
}
