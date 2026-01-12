package logic

import (
	"context"

	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchGroupLogic {
	return &SearchGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SearchGroup 搜索群组
func (l *SearchGroupLogic) SearchGroup(in *group.SearchGroupReq) (*group.SearchGroupResp, error) {
	// 使用 model 中定义的 SearchByKeyword 方法
	groups, err := l.svcCtx.ImGroupModel.SearchByKeyword(l.ctx, in.Keyword)
	if err != nil {
		logx.Errorf("搜索群组失败: %v", err)
		return nil, err
	}

	var list []*group.GroupInfo
	for _, g := range groups {
		list = append(list, &group.GroupInfo{
			Id:          int64(g.Id),
			GroupId:     g.GroupId,
			Name:        g.Name,
			Avatar:      g.Avatar.String,
			OwnerId:     g.OwnerId,
			Description: g.Description.String,
			MaxMembers:  int32(g.MaxMembers),
			MemberCount: int32(g.MemberCount),
			Status:      int32(g.Status),
			CreatedAt:   g.CreatedAt.Unix(),
			UpdatedAt:   g.UpdatedAt.Unix(),
		})
	}

	return &group.SearchGroupResp{
		Groups: list,
	}, nil
}
