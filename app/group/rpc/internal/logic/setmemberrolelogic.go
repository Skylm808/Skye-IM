package logic

import (
	"context"
	"errors"

	"SkyeIM/app/group/model"
	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetMemberRoleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetMemberRoleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetMemberRoleLogic {
	return &SetMemberRoleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SetMemberRole 设置成员角色 (管理员/普通成员)
func (l *SetMemberRoleLogic) SetMemberRole(in *group.SetMemberRoleReq) (*group.SetMemberRoleResp, error) {
	// 1. 验证群组
	groupInfo, err := l.svcCtx.ImGroupModel.FindOneByGroupId(l.ctx, in.GroupId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.New("群组不存在")
		}
		return nil, err
	}

	// 2. 验证操作者权限 (必须是群主)
	if groupInfo.OwnerId != in.OperatorId {
		return nil, errors.New("只有群主可以设置管理员")
	}

	// 3. 验证目标成员
	member, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.MemberId)
	if err != nil {
		return nil, errors.New("成员不存在")
	}

	// 4. 更新角色
	// Role: 1=Owner, 2=Admin, 3=Member
	member.Role = int64(in.Role)
	err = l.svcCtx.ImGroupMemberModel.Update(l.ctx, member)
	if err != nil {
		l.Logger.Errorf("设置角色失败: %v", err)
		return nil, errors.New("设置角色失败")
	}

	// 5. 推送通知
	_ = l.svcCtx.WsPushClient.PushGroupEvent(in.GroupId, "memberUpdate", map[string]interface{}{
		"userId": in.MemberId,
		"role":   in.Role,
	})

	return &group.SetMemberRoleResp{}, nil
}
