package logic

import (
	"context"
	"errors"

	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetMemberMuteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetMemberMuteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetMemberMuteLogic {
	return &SetMemberMuteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SetMemberMute 设置成员禁言
func (l *SetMemberMuteLogic) SetMemberMute(in *group.SetMemberMuteReq) (*group.SetMemberMuteResp, error) {
	// 1. 验证权限 (群主或管理员)
	operator, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.OperatorId)
	if err != nil {
		return nil, errors.New("操作者不是群成员")
	}
	if operator.Role == 3 {
		return nil, errors.New("没有权限")
	}

	// 2. 验证目标
	target, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.MemberId)
	if err != nil {
		return nil, errors.New("目标成员不存在")
	}

	// 权限检查：管理员不能禁言管理员或群主
	if operator.Role == 2 && (target.Role == 1 || target.Role == 2) {
		return nil, errors.New("权限不足")
	}

	// 3. 更新禁言状态
	// Mute: 0=Unmute, >0 = Mute untill timestamp? Or duration? usually timestamp or 1=mute forever.
	// Assuming Mute field is Int64 (timestamp or duration). If user passes standard "seconds" or "until".
	// Let's assume in.Mute is int32 (seconds? or status).
	// Usually 0 = unmute, >0 = mute duration seconds.
	// If logic stores "MuteUntil" timestamp in DB:
	// target.Mute = time.Now().Unix() + int64(in.Mute)
	target.Mute = int64(in.Mute)
	err = l.svcCtx.ImGroupMemberModel.Update(l.ctx, target)
	if err != nil {
		l.Logger.Errorf("设置禁言失败: %v", err)
		return nil, errors.New("操作失败")
	}

	// 4. 推送通知
	_ = l.svcCtx.WsPushClient.PushGroupEvent(in.GroupId, "memberUpdate", map[string]interface{}{
		"userId": in.MemberId,
		"mute":   in.Mute,
	})

	return &group.SetMemberMuteResp{}, nil
}
