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

type GetReceivedJoinRequestsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 获取所有管理群组的入群申请
func NewGetReceivedJoinRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReceivedJoinRequestsLogic {
	return &GetReceivedJoinRequestsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetReceivedJoinRequestsLogic) GetReceivedJoinRequests(req *types.GetJoinRequestsReq) (resp *types.Response, err error) {
	// 提取JWT中的用户ID
	userId, err := l.ctx.Value("userId").(json.Number).Int64()
	if err != nil {
		return &types.Response{
			Code:    401,
			Message: "未授权",
		}, nil
	}

	// 调用RPC获取所有管理群组的申请
	rpcResp, err := l.svcCtx.GroupRpc.GetAllManagedGroupJoinRequests(l.ctx, &group.GetAllManagedGroupJoinRequestsReq{
		OperatorId: userId,
		Page:       int64(req.Page),
		PageSize:   int64(req.PageSize),
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
