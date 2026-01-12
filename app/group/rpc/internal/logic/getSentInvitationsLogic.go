package logic

import (
	"context"
	"errors"

	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSentInvitationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSentInvitationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSentInvitationsLogic {
	return &GetSentInvitationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetSentInvitations 获取发出的群聊邀请列表
func (l *GetSentInvitationsLogic) GetSentInvitations(in *group.GetSentInvitationsReq) (*group.GetSentInvitationsResp, error) {
	// 1. 查询邀请人发出的邀请列表
	invitations, err := l.svcCtx.ImGroupInvitationModel.FindByInviterId(l.ctx, in.UserId, in.Page, in.PageSize)
	if err != nil {
		logx.Errorf("查询发出的邀请列表失败: %v", err)
		return nil, errors.New("查询失败")
	}

	// 2. 统计总数
	total, err := l.svcCtx.ImGroupInvitationModel.CountByInviterId(l.ctx, in.UserId)
	if err != nil {
		logx.Errorf("统计邀请数量失败: %v", err)
		total = 0
	}

	// 3. 转换为Proto格式
	list := make([]*group.GroupInvitationInfo, 0, len(invitations))
	for _, inv := range invitations {
		list = append(list, &group.GroupInvitationInfo{
			Id:        int64(inv.Id),
			GroupId:   inv.GroupId,
			InviterId: int64(inv.InviterId),
			InviteeId: int64(inv.InviteeId),
			Message:   inv.Message,
			Status:    int64(inv.Status),
			CreatedAt: inv.CreatedAt.Unix(),
		})
	}

	return &group.GetSentInvitationsResp{
		List:  list,
		Total: total,
	}, nil
}
