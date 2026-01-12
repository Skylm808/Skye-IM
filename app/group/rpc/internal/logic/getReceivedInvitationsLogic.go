package logic

import (
	"context"
	"errors"

	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReceivedInvitationsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetReceivedInvitationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReceivedInvitationsLogic {
	return &GetReceivedInvitationsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetReceivedInvitations 获取收到的群聊邀请列表
func (l *GetReceivedInvitationsLogic) GetReceivedInvitations(in *group.GetReceivedInvitationsReq) (*group.GetReceivedInvitationsResp, error) {
	// 1. 查询被邀请人收到的邀请列表
	invitations, err := l.svcCtx.ImGroupInvitationModel.FindByInviteeId(l.ctx, in.UserId, in.Page, in.PageSize)
	if err != nil {
		logx.Errorf("查询收到的邀请列表失败: %v", err)
		return nil, errors.New("查询失败")
	}

	// 2. 统计总数
	total, err := l.svcCtx.ImGroupInvitationModel.CountByInviteeId(l.ctx, in.UserId)
	if err != nil {
		logx.Errorf("统计邀请数量失败: %v", err)
		total = 0
	}

	// 3. 转换为Proto格式
	list := make([]*group.GroupInvitationInfo, 0, len(invitations))
	for _, inv := range invitations {
		// 注意：根据模型定义，UserId 通常为 uint64，Proto 中为 int64，需要转换
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

	return &group.GetReceivedInvitationsResp{
		List:  list,
		Total: total,
	}, nil
}
