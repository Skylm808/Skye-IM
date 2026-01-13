// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package joinrequest

import (
	"context"
	"encoding/json"

	"SkyeIM/app/group/api/internal/svc"
	"SkyeIM/app/group/api/internal/types"
	"SkyeIM/app/group/rpc/group"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleJoinRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewHandleJoinRequestLogic 处理入群申请
func NewHandleJoinRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleJoinRequestLogic {
	return &HandleJoinRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleJoinRequestLogic) HandleJoinRequest(req *types.HandleJoinRequestReq) (resp *types.Response, err error) {
	// 提取JWT中的用户ID
	userId, err := l.ctx.Value("userId").(json.Number).Int64()
	if err != nil {
		return &types.Response{
			Code:    401,
			Message: "未授权",
		}, nil
	}

	// 调用RPC
	_, err = l.svcCtx.GroupRpc.HandleJoinRequest(l.ctx, &group.HandleJoinRequestReq{
		OperatorId: userId,
		RequestId:  req.RequestId,
		Action:     req.Action,
	})

	if err != nil {
		return &types.Response{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	message := "已拒绝入群申请"
	if req.Action == 1 {
		message = "已同意入群申请"
	}

	return &types.Response{
		Code:    200,
		Message: message,
	}, nil
}
