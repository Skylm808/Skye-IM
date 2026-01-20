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

	// 6. 更新 Redis
	go func() {
		redisKey := fmt.Sprintf("im:group:members:%s", in.GroupId)
		_, _ = l.svcCtx.Redis.Srem(redisKey, in.MemberId)
	}()

	// 7. 推送通知
	_ = l.svcCtx.WsPushClient.PushGroupEvent(in.GroupId, "kickMember", map[string]interface{}{
		"operatorId": in.OperatorId,
		"memberId":   in.MemberId,
		"groupId":    in.GroupId,
	})

	return &group.KickMemberResp{}, nil
}
