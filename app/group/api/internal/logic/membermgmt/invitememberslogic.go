// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package membermgmt

import (
	"context"
	"encoding/json"
	"fmt"

	"SkyeIM/app/group/api/internal/svc"
	"SkyeIM/app/group/api/internal/types"
	"SkyeIM/app/group/rpc/groupclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type InviteMembersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewInviteMembersLogic creates a new InviteMembersLogic.
func NewInviteMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InviteMembersLogic {
	return &InviteMembersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InviteMembersLogic) InviteMembers(req *types.InviteMembersReq) (resp *types.Response, err error) {
	userId := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	uid, _ := userId.Int64()

	rpcRes, err := l.svcCtx.GroupRpc.InviteMembers(l.ctx, &groupclient.InviteMembersReq{
		GroupId:   req.GroupId,
		InviterId: uid,
		MemberIds: req.MemberIds,
	})
	if err != nil {
		return nil, err
	}

	return &types.Response{
		Code:    0,
		Message: "success",
		Data: types.InviteMembersResp{
			SuccessCount: rpcRes.SuccessCount,
			FailedIds:    rpcRes.FailedIds,
		},
	}, nil
}
