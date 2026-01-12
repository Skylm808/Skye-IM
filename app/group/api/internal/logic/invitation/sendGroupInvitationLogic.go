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

type SendGroupInvitationLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSendGroupInvitationLogic creates a new SendGroupInvitationLogic.
func NewSendGroupInvitationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendGroupInvitationLogic {
	return &SendGroupInvitationLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendGroupInvitationLogic) SendGroupInvitation(req *types.SendGroupInvitationReq) (resp *types.Response, err error) {
	userID := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	inviterID, err := userID.Int64()
	if err != nil {
		return &types.Response{
			Code:    400,
			Message: "invalid user id",
		}, nil
	}

	rpcResp, err := l.svcCtx.GroupRpc.SendGroupInvitation(l.ctx, &groupclient.SendGroupInvitationReq{
		GroupId:   req.GroupId,
		InviterId: inviterID,
		InviteeId: req.InviteeId,
		Message:   req.Message,
	})
	if err != nil {
		logx.Errorf("send group invitation failed: %v", err)
		return &types.Response{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    0,
		Message: "success",
		Data: types.SendGroupInvitationResp{
			InvitationId: rpcResp.InvitationId,
		},
	}, nil
}
