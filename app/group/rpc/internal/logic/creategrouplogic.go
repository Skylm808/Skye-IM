package logic

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"SkyeIM/app/group/model"
	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/group/rpc/internal/svc"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateGroupLogic {
	return &CreateGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建群组
func (l *CreateGroupLogic) CreateGroup(in *group.CreateGroupReq) (*group.CreateGroupResp, error) {
	// 验证参数
	if in.OwnerId == 0 {
		return nil, status.Error(codes.InvalidArgument, "群主ID不能为空")
	}
	if in.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "群名称不能为空")
	}

	// 设置默认值
	maxMembers := in.MaxMembers
	if maxMembers == 0 {
		maxMembers = 200
	}

	// 生成群组ID
	groupId := uuid.New().String()

	// 创建群组记录
	groupData := &model.ImGroup{
		GroupId:     groupId,
		Name:        in.Name,
		OwnerId:     in.OwnerId,
		MaxMembers:  int64(maxMembers),
		MemberCount: 1, // 创建者自动加入
		Status:      1, // 正常状态
	}

	// 设置可选字段
	if in.Avatar != "" {
		groupData.Avatar = sql.NullString{String: in.Avatar, Valid: true}
	}
	if in.Description != "" {
		groupData.Description = sql.NullString{String: in.Description, Valid: true}
	}

	// 插入群组
	result, err := l.svcCtx.ImGroupModel.Insert(l.ctx, groupData)
	if err != nil {
		l.Logger.Errorf("创建群组失败: %v", err)
		return nil, status.Error(codes.Internal, "创建群组失败")
	}

	groupDbId, _ := result.LastInsertId()

	// 将创建者添加为群主
	memberData := &model.ImGroupMember{
		GroupId:  groupId,
		UserId:   in.OwnerId,
		Role:     1, // 群主
		Mute:     0,
		JoinedAt: time.Now(),
	}

	_, err = l.svcCtx.ImGroupMemberModel.Insert(l.ctx, memberData)
	if err != nil {
		l.Logger.Errorf("添加群主失败: %v", err)
		// 回滚：删除刚创建的群组
		l.svcCtx.ImGroupModel.Delete(l.ctx, groupDbId)
		return nil, status.Error(codes.Internal, "添加群主失败")
	}

	// 如果有初始成员，批量添加
	if len(in.MemberIds) > 0 {
		successCount := 0
		for _, memberId := range in.MemberIds {
			if memberId == in.OwnerId {
				continue // 跳过群主
			}

			// 检查人数限制
			if int32(groupData.MemberCount)+int32(successCount) >= maxMembers {
				break
			}

			member := &model.ImGroupMember{
				GroupId:  groupId,
				UserId:   memberId,
				Role:     3, // 普通成员
				Mute:     0,
				JoinedAt: time.Now(),
			}

			_, err := l.svcCtx.ImGroupMemberModel.Insert(l.ctx, member)
			if err != nil {
				l.Logger.Errorf("添加成员 %d 失败: %v", memberId, err)
				continue
			}
			successCount++
		}

		// 更新群成员数
		if successCount > 0 {
			groupData.Id = groupDbId
			groupData.MemberCount = int64(1 + successCount)
			l.svcCtx.ImGroupModel.Update(l.ctx, groupData)
		}
	}

	// 6. 异步更新 Redis 缓存 (群成员列表)
	go func() {
		redisKey := fmt.Sprintf("im:group:members:%s", groupId)
		memberIds := []interface{}{in.OwnerId}
		for _, uid := range in.MemberIds {
			if uid != in.OwnerId {
				memberIds = append(memberIds, uid)
			}
		}
		if _, err := l.svcCtx.Redis.Sadd(redisKey, memberIds...); err != nil {
			l.Logger.Errorf("Redis 更新群成员失败: %v", err)
		}
		l.svcCtx.Redis.Expire(redisKey, 7*24*60*60)
	}()

	// 返回结果
	// 构造响应
	resp := &group.CreateGroupResp{
		Group: &group.GroupInfo{
			Id:          groupDbId,
			GroupId:     groupId,
			Name:        groupData.Name,
			Avatar:      groupData.Avatar.String,
			OwnerId:     groupData.OwnerId,
			Description: groupData.Description.String,
			MaxMembers:  int32(groupData.MaxMembers),
			MemberCount: int32(groupData.MemberCount),
			Status:      int32(groupData.Status),
			CreatedAt:   groupData.CreatedAt.Unix(),
			UpdatedAt:   groupData.UpdatedAt.Unix(),
		},
	}

	// 推送创建事件
	_ = l.svcCtx.WsPushClient.PushGroupEvent(groupId, "createGroup", map[string]interface{}{
		"operatorId": in.OwnerId,
		"group":      resp.Group,
	})

	return resp, nil
}
