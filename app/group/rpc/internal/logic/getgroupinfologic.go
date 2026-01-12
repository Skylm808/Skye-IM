package logic

import (
	"context"
	"errors"

	"SkyeIM/app/group/model"
	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupInfoLogic {
	return &GetGroupInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetGroupInfo 获取群组信息
func (l *GetGroupInfoLogic) GetGroupInfo(in *group.GetGroupInfoReq) (*group.GetGroupInfoResp, error) {
	groupInfo, err := l.svcCtx.ImGroupModel.FindOneByGroupId(l.ctx, in.GroupId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.New("群组不存在")
		}
		return nil, err
	}

	return &group.GetGroupInfoResp{
		Group: &group.GroupInfo{
			Id:          int64(groupInfo.Id),
			GroupId:     groupInfo.GroupId,
			Name:        groupInfo.Name,
			Avatar:      groupInfo.Avatar.String,
			OwnerId:     groupInfo.OwnerId,
			Description: groupInfo.Description.String,
			MaxMembers:  int32(groupInfo.MaxMembers),
			MemberCount: int32(groupInfo.MemberCount),
			Status:      int32(groupInfo.Status),
			CreatedAt:   groupInfo.CreatedAt.Unix(),
			UpdatedAt:   groupInfo.UpdatedAt.Unix(),
		},
	}, nil
}
