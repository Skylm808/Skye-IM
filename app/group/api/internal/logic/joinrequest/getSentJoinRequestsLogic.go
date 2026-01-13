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

type GetSentJoinRequestsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetSentJoinRequestsLogic 获取我发出的入群申请
func NewGetSentJoinRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSentJoinRequestsLogic {
	return &GetSentJoinRequestsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSentJoinRequestsLogic) GetSentJoinRequests(req *types.GetJoinRequestsReq) (resp *types.Response, err error) {
	// 提取JWT中的用户ID
	userId, err := l.ctx.Value("userId").(json.Number).Int64()
	if err != nil {
		return &types.Response{
			Code:    401,
			Message: "未授权",
		}, nil
	}

	// 调用RPC
	rpcResp, err := l.svcCtx.GroupRpc.GetSentJoinRequests(l.ctx, &group.GetSentJoinRequestsReq{
		UserId:   userId,
		Page:     int64(req.Page),
		PageSize: int64(req.PageSize),
	})

	if err != nil {
		return &types.Response{
			Code:    500,
			Message: err.Error(),
		}, nil
	}

	// 转换结果
	var list []types.JoinRequestInfo
	for _, item := range rpcResp.List {
		list = append(list, types.JoinRequestInfo{
			Id:        item.Id,
			GroupId:   item.GroupId,
			UserId:    item.UserId,
			Message:   item.Message,
			Status:    item.Status,
			HandlerId: item.HandlerId,
			CreatedAt: item.CreatedAt,
		})
	}

	return &types.Response{
		Code:    200,
		Message: "查询成功",
		Data: types.GetJoinRequestsResp{
			List:  list,
			Total: rpcResp.Total,
		},
	}, nil
}
