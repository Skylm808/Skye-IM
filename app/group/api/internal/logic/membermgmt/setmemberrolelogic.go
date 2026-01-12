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

type SetMemberRoleLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 设置成员角色
func NewSetMemberRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetMemberRoleLogic {
	return &SetMemberRoleLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetMemberRoleLogic) SetMemberRole(req *types.SetMemberRoleReq) (resp *types.Response, err error) {
	userId := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	uid, _ := userId.Int64()

	_, err = l.svcCtx.GroupRpc.SetMemberRole(l.ctx, &groupclient.SetMemberRoleReq{
		GroupId:    req.GroupId,
		OperatorId: uid,
		MemberId:   req.MemberId,
		Role:       req.Role,
	})

	if err != nil {
		return nil, err
	}

	return &types.Response{
		Code:    0,
		Message: "设置成功",
	}, nil
}
