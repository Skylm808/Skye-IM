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

type QuitGroupLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewQuitGroupLogic creates a new QuitGroupLogic.
func NewQuitGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QuitGroupLogic {
	return &QuitGroupLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QuitGroupLogic) QuitGroup(req *types.QuitGroupReq) (resp *types.Response, err error) {
	userId := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	uid, _ := userId.Int64()

	_, err = l.svcCtx.GroupRpc.QuitGroup(l.ctx, &groupclient.QuitGroupReq{
		GroupId: req.GroupId,
		UserId:  uid,
	})
	if err != nil {
		return nil, err
	}

	return &types.Response{
		Code:    0,
		Message: "success",
	}, nil
}
