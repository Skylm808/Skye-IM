// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package membermgmt

import (
	"SkyeIM/app/group/api/internal/svc"
	"SkyeIM/app/group/api/internal/types"
	"SkyeIM/app/group/rpc/groupclient"
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetMemberMuteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 设置成员禁言
func NewSetMemberMuteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetMemberMuteLogic {
	return &SetMemberMuteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetMemberMuteLogic) SetMemberMute(req *types.SetMemberMuteReq) (resp *types.Response, err error) {
	userId := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	uid, _ := userId.Int64()

	_, err = l.svcCtx.GroupRpc.SetMemberMute(l.ctx, &groupclient.SetMemberMuteReq{
		GroupId:    req.GroupId,
		OperatorId: uid,
		MemberId:   req.MemberId,
		Mute:       req.Mute,
	})

	if err != nil {
		return nil, err
	}

	return &types.Response{
		Code:    0,
		Message: "设置成功",
	}, nil
}
