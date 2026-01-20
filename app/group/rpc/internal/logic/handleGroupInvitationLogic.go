package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"SkyeIM/app/group/model"
	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleGroupInvitationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHandleGroupInvitationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleGroupInvitationLogic {
	return &HandleGroupInvitationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// HandleGroupInvitation 处理群聊邀请 (接受/拒绝)
func (l *HandleGroupInvitationLogic) HandleGroupInvitation(in *group.HandleGroupInvitationReq) (*group.HandleGroupInvitationResp, error) {
	// 1. 获取邀请记录
	invitation, err := l.svcCtx.ImGroupInvitationModel.FindOne(l.ctx, uint64(in.InvitationId))
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.New("邀请不存在")
		}
		return nil, err
	}

	// 2. 验证操作者是否为被邀请人
	if int64(invitation.InviteeId) != in.UserId {
		return nil, errors.New("无权处理此邀请")
	}

	// 3. 验证状态
	if invitation.Status != 0 {
		return nil, errors.New("邀请已处理")
	}

	// 4. 更新状态
	// in.Action: 1=Accept, 2=Reject
	invitation.Status = int64(in.Action)
	err = l.svcCtx.ImGroupInvitationModel.Update(l.ctx, invitation)
	if err != nil {
		l.Logger.Errorf("更新邀请状态失败: %v", err)
		return nil, errors.New("处理失败")
	}

	// 5. 如果接受邀请，添加成员
	if in.Action == 1 {
		// 检查是否已经是成员
		isMember, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, invitation.GroupId, int64(invitation.InviteeId))
		if err == nil && isMember != nil {
			return &group.HandleGroupInvitationResp{}, nil // 已经是成员，直接返回成功
		}

		// 获取群组信息并检查状态和成员上限
		groupInfo, err := l.svcCtx.ImGroupModel.FindOneByGroupId(l.ctx, invitation.GroupId)
		if err != nil {
			return nil, err
		}

		// 检查群组是否已解散
		if groupInfo.Status != 1 {
			return nil, errors.New("群组已解散，无法加入")
		}

		// 检查群成员是否已满
		if groupInfo.MemberCount >= groupInfo.MaxMembers {
			return nil, errors.New("群成员已满")
		}

		// 添加成员
		member := &model.ImGroupMember{
			GroupId:  invitation.GroupId,
			UserId:   int64(invitation.InviteeId),
			Role:     3, // 普通成员
			Mute:     0,
			JoinedAt: time.Now(),
		}
		_, err = l.svcCtx.ImGroupMemberModel.Insert(l.ctx, member)
		if err != nil {
			l.Logger.Errorf("添加成员失败: %v", err)
			return nil, errors.New("加入群组失败")
		}

		// 更新群人数
		groupInfo.MemberCount++
		l.svcCtx.ImGroupModel.Update(l.ctx, groupInfo)

		// 更新 Redis
		go func() {
			redisKey := fmt.Sprintf("im:group:members:%s", invitation.GroupId)
			l.svcCtx.Redis.Sadd(redisKey, invitation.InviteeId)
		}()

		// 推送入群通知
		l.svcCtx.WsPushClient.PushGroupEvent(invitation.GroupId, "joinGroup", map[string]interface{}{
			"userId":  invitation.InviteeId,
			"groupId": invitation.GroupId,
		})

		// 推送通知给邀请方
		go func() {
			_ = l.svcCtx.WsPushClient.PushToUser(int64(invitation.InviterId), "group_invitation_handled", map[string]interface{}{
				"invitationId": in.InvitationId,
				"groupId":      invitation.GroupId,
				"inviteeId":    invitation.InviteeId,
				"action":       "accepted",
				"handledAt":    time.Now().Unix(),
			})
		}()
	} else {
		// 拒绝邀请，推送通知给邀请方
		go func() {
			_ = l.svcCtx.WsPushClient.PushToUser(int64(invitation.InviterId), "group_invitation_handled", map[string]interface{}{
				"invitationId": in.InvitationId,
				"groupId":      invitation.GroupId,
				"inviteeId":    invitation.InviteeId,
				"action":       "rejected",
				"handledAt":    time.Now().Unix(),
			})
		}()
	}

	return &group.HandleGroupInvitationResp{}, nil
}
