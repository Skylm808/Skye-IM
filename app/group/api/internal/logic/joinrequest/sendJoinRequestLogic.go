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

type SendJoinRequestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSendJoinRequestLogic 发送入群申请
func NewSendJoinRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendJoinRequestLogic {
	return &SendJoinRequestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendJoinRequestLogic) SendJoinRequest(req *types.SendJoinRequestReq) (resp *types.Response, err error) {
	// 提取JWT中的用户ID
	userId, err := l.ctx.Value("userId").(json.Number).Int64()
	if err != nil {
		return &types.Response{
			Code:    401,
			Message: "未授权",
		}, nil
	}

	// 调用RPC
	rpcResp, err := l.svcCtx.GroupRpc.SendJoinRequest(l.ctx, &group.SendJoinRequestReq{
		GroupId: req.GroupId,
		UserId:  userId,
		Message: req.Message,
	})

	if err != nil {
		return &types.Response{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	return &types.Response{
		Code:    200,
		Message: "申请已发送",
		Data: types.SendJoinRequestResp{
			RequestId: rpcResp.RequestId,
		},
	}, nil
}
