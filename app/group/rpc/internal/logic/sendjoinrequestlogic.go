package logic

import (
	"context"
	"database/sql"
	"errors"

	"SkyeIM/app/group/model"
	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
)

type SendJoinRequestLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendJoinRequestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendJoinRequestLogic {
	return &SendJoinRequestLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SendJoinRequest 发送入群申请
func (l *SendJoinRequestLogic) SendJoinRequest(in *group.SendJoinRequestReq) (*group.SendJoinRequestResp, error) {
	// 1. 验证群组是否存在且状态正常
	groupInfo, err := l.svcCtx.ImGroupModel.FindOneByGroupId(l.ctx, in.GroupId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.New("群组不存在")
		}
		return nil, err
	}
	if groupInfo.Status != 1 {
		return nil, errors.New("群组已解散")
	}

	// 2. 检查用户是否已经是群组成员
	member, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.UserId)
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}
	if member != nil {
		return nil, errors.New("您已经是群成员")
	}

	// 3. 检查是否已有待处理的申请（防止重复申请）
	pendingRequest, err := l.svcCtx.ImGroupJoinRequestModel.FindPendingByGroupAndUser(l.ctx, in.GroupId, uint64(in.UserId))
	if err != nil && err != sqlc.ErrNotFound && err != model.ErrNotFound {
		return nil, err
	}
	if pendingRequest != nil {
		return nil, errors.New("已有待处理的入群申请，请耐心等待")
	}

	// 3.5 检查是否有历史申请记录（已同意/已拒绝）
	// 如果用户之前申请过但已被处理（同意后被踢出，或被拒绝），复用该记录
	existingRequest, err := l.svcCtx.ImGroupJoinRequestModel.FindLatestByGroupAndUser(l.ctx, in.GroupId, uint64(in.UserId))
	if err != nil && err != sqlc.ErrNotFound && err != model.ErrNotFound {
		return nil, err
	}

	if existingRequest != nil {
		// 复用已有记录，更新为新的申请
		existingRequest.Message = in.Message
		existingRequest.Status = 0                              // 重置为待处理
		existingRequest.HandlerId = sql.NullInt64{Valid: false} // 清空处理人
		err = l.svcCtx.ImGroupJoinRequestModel.Update(l.ctx, existingRequest)
		if err != nil {
			return nil, err
		}
		return &group.SendJoinRequestResp{
			RequestId: int64(existingRequest.Id),
		}, nil
	}

	// 4. 没有任何历史记录，创建新的入群申请记录
	joinRequest := &model.ImGroupJoinRequest{
		GroupId: in.GroupId,
		UserId:  uint64(in.UserId),
		Message: in.Message,
		Status:  0, // 0-待处理
	}

	result, err := l.svcCtx.ImGroupJoinRequestModel.Insert(l.ctx, joinRequest)
	if err != nil {
		return nil, err
	}

	requestId, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &group.SendJoinRequestResp{
		RequestId: requestId,
	}, nil
}
