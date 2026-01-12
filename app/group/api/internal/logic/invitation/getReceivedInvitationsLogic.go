// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package invitation

import (
	"context"
	"encoding/json"
	"fmt"

	"SkyeIM/app/group/api/internal/svc"
	"SkyeIM/app/group/api/internal/types"
	"SkyeIM/app/group/rpc/groupclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReceivedInvitationsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetReceivedInvitationsLogic creates a new GetReceivedInvitationsLogic.
func NewGetReceivedInvitationsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReceivedInvitationsLogic {
	return &GetReceivedInvitationsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetReceivedInvitations returns invitations received by the current user.
func (l *GetReceivedInvitationsLogic) GetReceivedInvitations(req *types.GetInvitationsReq) (resp *types.Response, err error) {
	userID := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	currentUserID, err := userID.Int64()
	if err != nil {
		return &types.Response{
			Code:    400,
			Message: "invalid user id",
		}, nil
	}

	rpcResp, err := l.svcCtx.GroupRpc.GetReceivedInvitations(l.ctx, &groupclient.GetReceivedInvitationsReq{
		UserId:   currentUserID,
		Page:     int64(req.Page),
		PageSize: int64(req.PageSize),
	})
	if err != nil {
		logx.Errorf("get received invitations failed: %v", err)
		return &types.Response{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	invitations := make([]types.GroupInvitationInfo, 0, len(rpcResp.List))
	for _, inv := range rpcResp.List {
		invitations = append(invitations, types.GroupInvitationInfo{
			Id:        inv.Id,
			GroupId:   inv.GroupId,
			InviterId: inv.InviterId,
			InviteeId: inv.InviteeId,
			Message:   inv.Message,
			Status:    inv.Status,
			CreatedAt: inv.CreatedAt,
		})
	}

	return &types.Response{
		Code:    0,
		Message: "success",
		Data: types.GetInvitationsResp{
			List:  invitations,
			Total: rpcResp.Total,
		},
	}, nil
}
