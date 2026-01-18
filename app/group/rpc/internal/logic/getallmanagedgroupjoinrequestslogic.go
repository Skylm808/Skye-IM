package logic

import (
	"context"

	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllManagedGroupJoinRequestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAllManagedGroupJoinRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllManagedGroupJoinRequestsLogic {
	return &GetAllManagedGroupJoinRequestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取所有管理群组的入群申请（通知中心）
func (l *GetAllManagedGroupJoinRequestsLogic) GetAllManagedGroupJoinRequests(in *group.GetAllManagedGroupJoinRequestsReq) (*group.GetAllManagedGroupJoinRequestsResp, error) {
	// 1. 查询用户作为管理员/群主的所有群组ID
	groupIds, err := l.svcCtx.ImGroupMemberModel.FindManagedGroupsByUserId(l.ctx, in.OperatorId)
	if err != nil {
		return nil, err
	}

	// 2. 如果用户不是任何群的管理员，直接返回空列表
	if len(groupIds) == 0 {
		return &group.GetAllManagedGroupJoinRequestsResp{
			List:  []*group.JoinRequestInfo{},
			Total: 0,
		}, nil
	}

	// 3. 查询这些群组的所有待处理申请（status=0）
	requests, total, err := l.svcCtx.ImGroupJoinRequestModel.FindByGroupIdsAndStatus(l.ctx, groupIds, 0, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}

	// 4. 转换为Proto格式
	var list []*group.JoinRequestInfo
	for _, req := range requests {
		handlerId := int64(0)
		if req.HandlerId.Valid {
			handlerId = req.HandlerId.Int64
		}
		list = append(list, &group.JoinRequestInfo{
			Id:        int64(req.Id),
			GroupId:   req.GroupId,
			UserId:    int64(req.UserId),
			Message:   req.Message,
			Status:    req.Status,
			HandlerId: handlerId,
			CreatedAt: req.CreatedAt.Unix(),
		})
	}

	return &group.GetAllManagedGroupJoinRequestsResp{
		List:  list,
		Total: total,
	}, nil
}
