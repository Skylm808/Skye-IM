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

type UpdateGroupReadSeqLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 更新群组已读进度
func NewUpdateGroupReadSeqLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGroupReadSeqLogic {
	return &UpdateGroupReadSeqLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateGroupReadSeqLogic) UpdateGroupReadSeq(req *types.UpdateGroupReadSeqReq) (resp *types.Response, err error) {
	userId := json.Number(fmt.Sprintf("%v", l.ctx.Value("userId")))
	uid, _ := userId.Int64()

	_, err = l.svcCtx.GroupRpc.UpdateGroupReadSeq(l.ctx, &groupclient.UpdateGroupReadSeqReq{
		GroupId: req.GroupId,
		UserId:  uid,
		ReadSeq: req.ReadSeq,
	})

	if err != nil {
		return nil, err
	}

	return &types.Response{
		Code:    0,
		Message: "更新已读进度成功",
	}, nil
}
