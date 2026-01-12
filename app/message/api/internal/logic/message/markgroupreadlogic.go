// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package message

import (
	"context"
	"errors"

	"SkyeIM/app/group/rpc/groupclient"
	"SkyeIM/app/message/api/internal/svc"
	"SkyeIM/app/message/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkGroupReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 群聊已读上报（更新readSeq）
func NewMarkGroupReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkGroupReadLogic {
	return &MarkGroupReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *MarkGroupReadLogic) MarkGroupRead(req *types.MarkGroupReadReq) (resp *types.MarkGroupReadResp, err error) {
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	if req.GroupId == "" {
		return nil, errors.New("groupId is required")
	}

	rpcResp, err := l.svcCtx.GroupRpc.UpdateGroupReadSeq(l.ctx, &groupclient.UpdateGroupReadSeqReq{
		GroupId: req.GroupId,
		UserId:  userId,
		ReadSeq: req.ReadSeq,
	})
	if err != nil {
		l.Logger.Errorf("UpdateGroupReadSeq RPC failed: %v", err)
		return nil, err
	}

	return &types.MarkGroupReadResp{Success: rpcResp.Success}, nil
}
