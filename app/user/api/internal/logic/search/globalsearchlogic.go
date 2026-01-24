// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package search

import (
	"context"
	"strings"

	"SkyeIM/app/user/api/internal/svc"
	"SkyeIM/app/user/api/internal/types"
	"SkyeIM/app/user/rpc/userClient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GlobalSearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 全局模糊搜索（用户/群组）
func NewGlobalSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GlobalSearchLogic {
	return &GlobalSearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GlobalSearchLogic) GlobalSearch(req *types.GlobalSearchRequest) (resp *types.GlobalSearchResponse, err error) {
	keyword := strings.TrimSpace(req.Keyword)
	if keyword == "" {
		return &types.GlobalSearchResponse{}, nil
	}

	// 1. 通过RPC模糊搜索用户
	userResp, err := l.svcCtx.UserRpc.SearchUsersByKeyword(l.ctx, &userClient.SearchUsersByKeywordRequest{
		Keyword: keyword,
	})
	if err != nil {
		l.Logger.Errorf("RPC全局搜索用户失败: %v", err)
	}

	// 2. 模糊搜索群组（暂时保留直接数据库访问）
	groups, err := l.svcCtx.GroupModel.SearchByKeyword(l.ctx, keyword)
	if err != nil {
		l.Logger.Errorf("全局搜索群组失败: %v", err)
	}

	// 转换用户结果
	var userList []types.UserInfo
	if userResp != nil {
		for _, u := range userResp.Users {
			userList = append(userList, convertToUserInfo(u))
		}
	}

	// 转换群组结果
	var groupList []types.GroupInfo
	for _, g := range groups {
		groupList = append(groupList, types.GroupInfo{
			GroupId:     g.GroupId,
			Name:        g.Name,
			Avatar:      g.Avatar.String,
			Description: g.Description.String,
		})
	}

	return &types.GlobalSearchResponse{
		Users:  userList,
		Groups: groupList,
	}, nil
}
