package logic

import (
	"context"
	"errors"
	"fmt"

	"SkyeIM/app/group/model"
	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type KickMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewKickMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KickMemberLogic {
	return &KickMemberLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// KickMember 踢出群成员
func (l *KickMemberLogic) KickMember(in *group.KickMemberReq) (*group.KickMemberResp, error) {
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
		return nil, errors.New("只有群主可以踢人")
	}

	// 3. 验证被踢者
	if in.OperatorId == in.MemberId {
		return nil, errors.New("不能踢出自己")
	}

	// 4. 执行删除
	err = l.svcCtx.ImGroupMemberModel.DeleteByGroupIdUserId(l.ctx, in.GroupId, in.MemberId)
	if err != nil {
		l.Logger.Errorf("踢出成员失败: %v", err)
		return nil, errors.New("踢出成员失败")
	}

	// 5. 更新群人数
	groupInfo.MemberCount--
	_ = l.svcCtx.ImGroupModel.Update(l.ctx, groupInfo)

	// 6．更新 Redis 缓存
	go func() {
		// 6.1 删除手动管理的成员列表缓存
		redisKey := fmt.Sprintf("im:group:members:%s", in.GroupId)
		_, _ = l.svcCtx.Redis.Srem(redisKey, in.MemberId)

		// 6.2 删除群成员 Model 缓存（go-zero 自动生成）
		// 注意：DeleteByGroupIdUserId 已经自动删除了，但为了确保完全清理，这里再次删除
		memberCacheKey := fmt.Sprintf("cache:imGroupMember:groupId:userId:%s:%d", in.GroupId, in.MemberId)
		_, _ = l.svcCtx.Redis.Del(memberCacheKey)

		// 6.3 更新群组信息缓存（因为 MemberCount 变了）
		// Update 操作会自动删除群组缓存，下次查询时自动回填

		l.Logger.Infof("[KickMember] Cleared cache for member %d in group %s", in.MemberId, in.GroupId)
	}()

	// 7. 推送通知
	_ = l.svcCtx.WsPushClient.PushGroupEvent(in.GroupId, "kickMember", map[string]interface{}{
		"operatorId": in.OperatorId,
		"memberId":   in.MemberId,
		"groupId":    in.GroupId,
	})

	return &group.KickMemberResp{}, nil
}
