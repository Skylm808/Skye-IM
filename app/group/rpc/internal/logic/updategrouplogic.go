package logic

import (
	"context"
	"database/sql"
	"errors"

	"SkyeIM/app/group/model"
	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateGroupLogic {
	return &UpdateGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateGroup 更新群组信息
func (l *UpdateGroupLogic) UpdateGroup(in *group.UpdateGroupReq) (*group.UpdateGroupResp, error) {
	// 1. 验证群组
	groupInfo, err := l.svcCtx.ImGroupModel.FindOneByGroupId(l.ctx, in.GroupId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.New("群组不存在")
		}
		return nil, err
	}

	// 2. 验证权限 (群主或管理员?) 通常只有群主或管理员可以修改
	if groupInfo.OwnerId != in.OperatorId {
		// Administrator check logic todo...
		return nil, errors.New("无权修改群信息")
	}

	// 3. 更新字段
	if in.Name != "" {
		groupInfo.Name = in.Name
	}
	if in.Avatar != "" {
		groupInfo.Avatar = sql.NullString{String: in.Avatar, Valid: true}
	}
	if in.Description != "" {
		groupInfo.Description = sql.NullString{String: in.Description, Valid: true}
	}

	err = l.svcCtx.ImGroupModel.Update(l.ctx, groupInfo)
	if err != nil {
		l.Logger.Errorf("更新群组失败: %v", err)
		return nil, errors.New("更新失败")
	}

	// 4. 推送更新事件
	// ... (push logic)
	_ = l.svcCtx.WsPushClient.PushGroupEvent(in.GroupId, "updateGroup", map[string]interface{}{
		"groupId": in.GroupId,
		"name":    groupInfo.Name,
		// ... other fields
	})

	return &group.UpdateGroupResp{}, nil
}
