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

	// 4. 清除 Redis 缓存
	go func() {
		redisKey := fmt.Sprintf("im:group:members:%s", in.GroupId)
		_, _ = l.svcCtx.Redis.Del(redisKey)
	}()

	// 5. 推送通知
	_ = l.svcCtx.WsPushClient.PushGroupEvent(in.GroupId, "dismissGroup", map[string]interface{}{
		"groupId":    in.GroupId,
		"operatorId": in.OperatorId,
	})

	return &group.DismissGroupResp{}, nil
}
