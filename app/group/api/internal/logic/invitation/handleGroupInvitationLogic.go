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

type HandleGroupInvitationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewHandleGroupInvitationLogic creates a new HandleGroupInvitationLogic.
func NewHandleGroupInvitationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleGroupInvitationLogic {
	return &HandleGroupInvitationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// HandleGroupInvitation processes an invitation action (accept/reject).
func (l *HandleGroupInvitationLogic) HandleGroupInvitation(req *types.HandleGroupInvitationReq) (resp *types.Response, err error) {
	userID := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	currentUserID, err := userID.Int64()
	if err != nil {
		return &types.Response{
			Code:    400,
			Message: "invalid user id",
		}, nil
	}

	_, err = l.svcCtx.GroupRpc.HandleGroupInvitation(l.ctx, &groupclient.HandleGroupInvitationReq{
		UserId:       currentUserID,
		InvitationId: req.InvitationId,
		Action:       req.Action,
	})
	if err != nil {
		logx.Errorf("handle group invitation failed: %v", err)
		return &types.Response{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    0,
		Message: "success",
	}, nil
}
