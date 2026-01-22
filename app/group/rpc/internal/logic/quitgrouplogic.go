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

type QuitGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQuitGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QuitGroupLogic {
	return &QuitGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QuitGroup 退出群组
func (l *QuitGroupLogic) QuitGroup(in *group.QuitGroupReq) (*group.QuitGroupResp, error) {
	// 1. 验证群组
	groupInfo, err := l.svcCtx.ImGroupModel.FindOneByGroupId(l.ctx, in.GroupId)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.New("群组不存在")
		}
		return nil, err
	}

	// 2. 验证是否为群主 (群主不能直接退群，需先转让或解散)
	if groupInfo.OwnerId == in.UserId {
		return nil, errors.New("群主不能直接退群，请先转让群主或解散群组")
	}

	// 3. 验证是否为成员
	member, err := l.svcCtx.ImGroupMemberModel.FindOneByGroupIdUserId(l.ctx, in.GroupId, in.UserId)
	if err != nil {
		return nil, errors.New("你不是群成员")
	}

	// 4. 删除成员记录
	err = l.svcCtx.ImGroupMemberModel.Delete(l.ctx, member.Id)
	if err != nil {
		l.Logger.Errorf("退出群组失败: %v", err)
		return nil, errors.New("退出群组失败")
	}

	// 5. 更新群人数
	groupInfo.MemberCount--
	_ = l.svcCtx.ImGroupModel.Update(l.ctx, groupInfo)

	// 6. 更新 Redis 缓存
	go func() {
		// 6.1 删除手动管理的成员列表缓存
		redisKey := fmt.Sprintf("im:group:members:%s", in.GroupId)
		_, _ = l.svcCtx.Redis.Srem(redisKey, in.UserId)

		// 6.2 删除群成员 Model 缓存（go-zero 自动生成）
		// 注意：Delete 已经自动删除了，但为了确保完全清理，这里再次删除
		memberCacheKey := fmt.Sprintf("cache:imGroupMember:groupId:userId:%s:%d", in.GroupId, in.UserId)
		memberIdKey := fmt.Sprintf("cache:imGroupMember:id:%d", member.Id)
		_, _ = l.svcCtx.Redis.Del(memberCacheKey)
		_, _ = l.svcCtx.Redis.Del(memberIdKey)

		// 6.3 更新群组信息缓存（因为 MemberCount 变了）
		// Update 操作会自动删除群组缓存，下次查询时自动回填

		l.Logger.Infof("[QuitGroup] Cleared cache for user %d in group %s", in.UserId, in.GroupId)
	}()

	// 7. 推送通知
	_ = l.svcCtx.WsPushClient.PushGroupEvent(in.GroupId, "quitGroup", map[string]interface{}{
		"userId":  in.UserId,
		"groupId": in.GroupId,
	})

	return &group.QuitGroupResp{}, nil
}
