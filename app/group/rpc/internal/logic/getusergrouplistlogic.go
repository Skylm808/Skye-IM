package logic

import (
	"context"

	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserGroupListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserGroupListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserGroupListLogic {
	return &GetUserGroupListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetUserGroupList 获取用户加入的群组列表
func (l *GetUserGroupListLogic) GetUserGroupList(in *group.GetUserGroupListReq) (*group.GetUserGroupListResp, error) {
	// 使用 model 中定义的 FindGroupsByUserId 方法 (支持分页)
	members, err := l.svcCtx.ImGroupMemberModel.FindGroupsByUserId(l.ctx, in.UserId, in.Page, in.PageSize)
	if err != nil {
		logx.Errorf("查询用户群组失败: %v", err)
		return nil, err
	}

	var groupList []*group.GroupInfo
	for _, m := range members {
		// 2. 查询和组装群信息
		g, err := l.svcCtx.ImGroupModel.FindOneByGroupId(l.ctx, m.GroupId)
		if err != nil {
			continue
		}
		groupList = append(groupList, &group.GroupInfo{
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

	// 3. 统计总数
	total, _ := l.svcCtx.ImGroupMemberModel.CountByUserId(l.ctx, in.UserId)

	return &group.GetUserGroupListResp{
		Groups: groupList,
		Total:  int32(total),
	}, nil
}
