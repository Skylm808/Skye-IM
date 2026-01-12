// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package groupmgmt

import (
	"context"

	"SkyeIM/app/group/api/internal/svc"
	"SkyeIM/app/group/api/internal/types"
	"SkyeIM/app/group/rpc/groupclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 模糊搜索群组
func NewSearchGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchGroupLogic {
	return &SearchGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchGroupLogic) SearchGroup(req *types.SearchGroupReq) (resp *types.Response, err error) {
	rpcRes, err := l.svcCtx.GroupRpc.SearchGroup(l.ctx, &groupclient.SearchGroupReq{
		Keyword: req.Keyword,
	})
	if err != nil {
		l.Logger.Errorf("API 搜索群组失败: %v", err)
		return nil, err
	}

	var list []types.GroupInfo
	for _, v := range rpcRes.Groups {
		list = append(list, types.GroupInfo{
			GroupId:     v.GroupId,
			Name:        v.Name,
			Avatar:      v.Avatar,
			OwnerId:     v.OwnerId,
			Description: v.Description,
			MaxMembers:  v.MaxMembers,
			MemberCount: v.MemberCount,
			Status:      v.Status,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		})
	}

	return &types.Response{
		Code:    0,
		Message: "获取成功",
		Data: types.SearchGroupResp{
			List:  list,
			Total: int64(len(list)),
		},
	}, nil
}
