package logic

import (
	"context"
	"errors"

	"SkyeIM/app/group/model"
	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGroupJoinRequestsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetGroupJoinRequestsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGroupJoinRequestsLogic {
	return &GetGroupJoinRequestsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetGroupJoinRequests 获取群的入群申请列表（群主/管理员查看）
func (l *GetGroupJoinRequestsLogic) GetGroupJoinRequests(in *group.GetGroupJoinRequestsReq) (*group.GetGroupJoinRequestsResp, error) {
	// 1. 验证操作者权限（必须是群主或管理员）
	member, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.OperatorId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.New("您不是群成员，无权查看申请")
		}
		return nil, err
	}
	if member.Role != 1 && member.Role != 2 {
		return nil, errors.New("只有群主或管理员可以查看申请列表")
	}

	// 2. 分页查询待处理的申请（status=0）
	requests, err := l.svcCtx.ImGroupJoinRequestModel.FindByGroupIdAndStatus(l.ctx, in.GroupId, 0, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}

	// 3. 统计总数
	total, err := l.svcCtx.ImGroupJoinRequestModel.CountByGroupIdAndStatus(l.ctx, in.GroupId, 0)
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

	return &group.GetGroupJoinRequestsResp{
		List:  list,
		Total: total,
	}, nil
}
