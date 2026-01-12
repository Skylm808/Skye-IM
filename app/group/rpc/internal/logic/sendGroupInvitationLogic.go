package logic

import (
	"context"
	"errors"
	"time"

	"SkyeIM/app/group/model"
	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendGroupInvitationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendGroupInvitationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendGroupInvitationLogic {
	return &SendGroupInvitationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ==================== 群聊邀请相关 ====================
// SendGroupInvitation 发送群聊邀请
func (l *SendGroupInvitationLogic) SendGroupInvitation(in *group.SendGroupInvitationReq) (*group.SendGroupInvitationResp, error) {
	// 1. 验证群组是否存在且状态正常
	groupInfo, err := l.svcCtx.ImGroupModel.FindOneByGroupId(l.ctx, in.GroupId)
	if err != nil {
		logx.Errorf("群组不存在或已解散: %v", err)
		return nil, errors.New("群组不存在")
	}
	if groupInfo.Status != 1 {
		return nil, errors.New("群组已解散")
	}

	// 2. 验证邀请人是否为群成员
	inviterMember, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.InviterId)
	if err != nil {
		logx.Errorf("邀请人不是群成员: %v", err)
		return nil, errors.New("您不是群成员，无法邀请")
	}
	if inviterMember == nil {
		return nil, errors.New("您不是群成员，无法邀请")
	}

	// 3. 验证被邀请人是否已在群中
	existingMember, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.InviteeId)
	if err == nil && existingMember != nil {
		return nil, errors.New("该用户已在群中")
	}

	// 4. 检查是否存在待处理的邀请（避免重复邀请）
	pendingInvitation, err := l.svcCtx.ImGroupInvitationModel.FindPendingByGroupAndInvitee(l.ctx, in.GroupId, in.InviteeId)
	if err == nil && pendingInvitation != nil {
		return nil, errors.New("已存在待处理的邀请")
	}

	// 5. 创建邀请记录
	invitation := &model.ImGroupInvitation{
		GroupId:   in.GroupId,
		InviterId: uint64(in.InviterId),
		InviteeId: uint64(in.InviteeId),
		Message:   in.Message,
		Status:    0, // 待处理
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := l.svcCtx.ImGroupInvitationModel.Insert(l.ctx, invitation)
	if err != nil {
		logx.Errorf("创建邀请记录失败: %v", err)
		return nil, errors.New("发送邀请失败")
	}

	invitationId, err := result.LastInsertId()
	if err != nil {
		logx.Errorf("获取邀请ID失败: %v", err)
		return nil, errors.New("发送邀请失败")
	}

	// 6. 通过WebSocket推送通知给被邀请人（后续补充PushToUser方法）
	// TODO: 实现WebSocket推送
	/*
		go l.svcCtx.WsPushClient.PushToUser(in.InviteeId, map[string]interface{}{
			"type": 9, // MessageTypeGroupInvitation
			"data": map[string]interface{}{
				"invitationId": invitationId,
				"groupId":      in.GroupId,
				"groupName":    groupInfo.Name,
				"inviterId":    in.InviterId,
				"message":      in.Message,
				"createdAt":    time.Now().Unix(),
			},
		})
	*/

	return &group.SendGroupInvitationResp{
		InvitationId: invitationId,
	}, nil
}
