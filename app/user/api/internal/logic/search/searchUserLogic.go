// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package search

import (
	"context"
	"fmt"
	"strings"

	"SkyeIM/app/user/api/internal/svc"
	"SkyeIM/app/user/api/internal/types"
	"SkyeIM/app/user/rpc/userClient"

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

	// 通过RPC精确搜索用户
	searchResp, err := l.svcCtx.UserRpc.SearchUser(l.ctx, &userClient.SearchUserRequest{
		Keyword: keyword,
	})
	if err != nil {
		l.Logger.Errorf("RPC搜索用户失败: %v", err)
		return nil, fmt.Errorf("搜索用户失败")
	}

	// 转换结果
	var users []types.UserInfo
	for _, u := range searchResp.Users {
		users = append(users, convertToUserInfo(u))
	}

	return &types.SearchUserResponse{
		Users: users,
		Total: searchResp.Total,
	}, nil
}
