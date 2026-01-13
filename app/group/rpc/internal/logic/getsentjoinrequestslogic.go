package logic

import (
	"context"

	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSentJoinRequestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSentJoinRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSentJoinRequestsLogic {
	return &GetSentJoinRequestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetSentJoinRequests 获取用户发出的入群申请
func (l *GetSentJoinRequestsLogic) GetSentJoinRequests(in *group.GetSentJoinRequestsReq) (*group.GetSentJoinRequestsResp, error) {
	// 1. 分页查询用户的申请列表
	requests, err := l.svcCtx.ImGroupJoinRequestModel.FindByUserId(l.ctx, uint64(in.UserId), in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}

	// 2. 统计总数
	total, err := l.svcCtx.ImGroupJoinRequestModel.CountByUserId(l.ctx, uint64(in.UserId))
	if err != nil {
		return nil, err
	}

	// 3. 转换为Proto格式
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

	return &group.GetSentJoinRequestsResp{
		List:  list,
		Total: total,
	}, nil
}
