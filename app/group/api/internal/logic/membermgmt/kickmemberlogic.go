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

type KickMemberLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 踢出成员
func NewKickMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KickMemberLogic {
	return &KickMemberLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *KickMemberLogic) KickMember(req *types.KickMemberReq) (resp *types.Response, err error) {
	userId := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	uid, _ := userId.Int64()

	_, err = l.svcCtx.GroupRpc.KickMember(l.ctx, &groupclient.KickMemberReq{
		GroupId:    req.GroupId,
		OperatorId: uid,
		MemberId:   req.MemberId,
	})

	if err != nil {
		return nil, err
	}

	return &types.Response{
		Code:    0,
		Message: "踢出成功",
	}, nil
}
