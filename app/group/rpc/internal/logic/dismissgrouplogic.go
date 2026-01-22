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

type DismissGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDismissGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DismissGroupLogic {
	return &DismissGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DismissGroup 解散群组
func (l *DismissGroupLogic) DismissGroup(in *group.DismissGroupReq) (*group.DismissGroupResp, error) {
	// 1. 验证群组是否存在
	groupInfo, err := l.svcCtx.ImGroupModel.FindOneByGroupId(l.ctx, in.GroupId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.New("群组不存在")
		}
		return nil, err
	}

	// 2. 验证是否为群主
	if groupInfo.OwnerId != in.OperatorId {
		return nil, errors.New("只有群主可以解散群组")
	}

	// 3. 更新状态为已解散 (Status = 2)
	groupInfo.Status = 2
	err = l.svcCtx.ImGroupModel.Update(l.ctx, groupInfo)
	if err != nil {
		l.Logger.Errorf("解散群组失败: %v", err)
		return nil, errors.New("解散群组失败")
	}

	// 3.5 删除所有群成员记录（同步执行，确保数据一致性）
	members, err := l.svcCtx.ImGroupMemberModel.FindByGroupId(l.ctx, in.GroupId)
	if err == nil && len(members) > 0 {
		l.Logger.Infof("[DismissGroup] Starting to delete %d members for group %s", len(members), in.GroupId)

		for _, member := range members {
			err := l.svcCtx.ImGroupMemberModel.Delete(l.ctx, member.Id)
			if err != nil {
				l.Logger.Errorf("[DismissGroup] Failed to delete member %d: %v", member.Id, err)
				// 继续删除其他成员，不中断流程
			}
		}

		l.Logger.Infof("[DismissGroup] Successfully deleted %d members", len(members))
	} else if err != nil {
		l.Logger.Errorf("[DismissGroup] Failed to query members: %v", err)
	}

	// 4. 清除 Redis 缓存
	go func() {
		// 4.1 删除手动管理的缓存
		memberKey := fmt.Sprintf("im:group:members:%s", in.GroupId)
		seqKey := fmt.Sprintf("group:seq:%s", in.GroupId)
		_, _ = l.svcCtx.Redis.Del(memberKey)
		_, _ = l.svcCtx.Redis.Del(seqKey)

		// 4.2 删除群组信息缓存（go-zero Model 层）
		// 注意：Update 操作已经自动删除了这些缓存，但为了确保完全清理，这里再次删除
		groupIdCacheKey := fmt.Sprintf("cache:imAuth:imGroup:groupId:%s", in.GroupId)
		groupPrimaryKey := fmt.Sprintf("cache:imAuth:imGroup:id:%d", groupInfo.Id)
		_, _ = l.svcCtx.Redis.Del(groupIdCacheKey)
		_, _ = l.svcCtx.Redis.Del(groupPrimaryKey)

		// 4.3 批量删除群成员缓存（go-zero Model 层）
		// 使用 SCAN 代替 KEYS 命令，避免阻塞 Redis
		pattern := fmt.Sprintf("cache:imGroupMember:groupId:userId:%s:*", in.GroupId)
		cursor := uint64(0)
		for {
			keys, nextCursor, err := l.svcCtx.Redis.Scan(cursor, pattern, 100)
			if err != nil {
				break
			}
			if len(keys) > 0 {
				_, _ = l.svcCtx.Redis.Del(keys...)
			}
			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}

		l.Logger.Infof("[DismissGroup] Cleared all cache for group %s", in.GroupId)
	}()

	// 5. 推送通知
	_ = l.svcCtx.WsPushClient.PushGroupEvent(in.GroupId, "dismissGroup", map[string]interface{}{
		"groupId":    in.GroupId,
		"operatorId": in.OperatorId,
	})

	return &group.DismissGroupResp{}, nil
}
