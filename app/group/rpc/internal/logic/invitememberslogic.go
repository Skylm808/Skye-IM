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

type InviteMembersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewInviteMembersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InviteMembersLogic {
	return &InviteMembersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// InviteMembers 邀请成员加入群组 (批量直接加入)
func (l *InviteMembersLogic) InviteMembers(in *group.InviteMembersReq) (*group.InviteMembersResp, error) {
	// 1. 验证群组
	groupInfo, err := l.svcCtx.ImGroupModel.FindOneByGroupId(l.ctx, in.GroupId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.New("群组不存在")
		}
		return nil, err
	}

	// 2. 验证邀请人权限
	_, err = l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.InviterId)
	if err != nil {
		return nil, errors.New("邀请人不是群成员")
	}

	var successIds []int64
	var failedIds []int64

	// 3. 批量处理
	for _, inviteeId := range in.MemberIds {
		// 检查是否已经是成员
		isMember, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, inviteeId)
		if err == nil && isMember != nil {
			failedIds = append(failedIds, inviteeId)
			continue
		}

		// 检查群人数上限
		if groupInfo.MemberCount >= groupInfo.MaxMembers {
			failedIds = append(failedIds, inviteeId)
			continue
		}

		// 添加成员
		member := &model.ImGroupMember{
			GroupId:  in.GroupId,
			UserId:   inviteeId,
			Role:     3, // 普通成员
			Mute:     0,
			JoinedAt: time.Now(),
		}
		_, err = l.svcCtx.ImGroupMemberModel.Insert(l.ctx, member)
		if err != nil {
			logx.Errorf("添加成员失败: %v", err)
			failedIds = append(failedIds, inviteeId)
			continue
		}

		successIds = append(successIds, inviteeId)
		// 更新内存中的计数，最后一次性更新数据库
		groupInfo.MemberCount++
	}

	// 4. 更新群组信息及缓存
	if len(successIds) > 0 {
		_ = l.svcCtx.ImGroupModel.Update(l.ctx, groupInfo)

		// Update Redis
		go func() {
			redisKey := fmt.Sprintf("im:group:members:%s", in.GroupId)
			ids := make([]interface{}, len(successIds))
			for i, id := range successIds {
				ids[i] = id
			}
			_, _ = l.svcCtx.Redis.Sadd(redisKey, ids...)
		}()

		// 5. 推送通知
		_ = l.svcCtx.WsPushClient.PushGroupEvent(in.GroupId, "inviteMember", map[string]interface{}{
			"inviterId": in.InviterId,
			"memberIds": successIds,
		})
	}

	return &group.InviteMembersResp{
		SuccessCount: int32(len(successIds)),
		FailedIds:    failedIds,
	}, nil
}
