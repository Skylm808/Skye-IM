package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImGroupJoinRequestModel = (*customImGroupJoinRequestModel)(nil)

type (
	// ImGroupJoinRequestModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImGroupJoinRequestModel.
	ImGroupJoinRequestModel interface {
		imGroupJoinRequestModel
		// 检查是否有待处理的申请
		FindPendingByGroupAndUser(ctx context.Context, groupId string, userId int64) (*ImGroupJoinRequest, error)
		// 按群组和状态查询申请列表
		FindByGroupAndStatus(ctx context.Context, groupId string, status int64, page, pageSize int64) ([]*ImGroupJoinRequest, error)
		// 按群组查询所有申请（分页）
		FindByGroup(ctx context.Context, groupId string, page, pageSize int64) ([]*ImGroupJoinRequest, error)
		// 查询用户发出的申请
		FindByUserId(ctx context.Context, userId int64, page, pageSize int64) ([]*ImGroupJoinRequest, error)
		// 统计群待处理申请数
		CountPendingByGroup(ctx context.Context, groupId string) (int64, error)
		// 统计用户发出的申请数
		CountByUserId(ctx context.Context, userId int64) (int64, error)
	}

	customImGroupJoinRequestModel struct {
		*defaultImGroupJoinRequestModel
	}
)

// NewImGroupJoinRequestModel returns a model for the database table.
func NewImGroupJoinRequestModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ImGroupJoinRequestModel {
	return &customImGroupJoinRequestModel{
		defaultImGroupJoinRequestModel: newImGroupJoinRequestModel(conn, c, opts...),
	}
}

// FindPendingByGroupAndUser 检查是否有待处理的申请
func (m *customImGroupJoinRequestModel) FindPendingByGroupAndUser(ctx context.Context, groupId string, userId int64) (*ImGroupJoinRequest, error) {
	var request ImGroupJoinRequest
	query := "SELECT * FROM im_group_join_request WHERE group_id = ? AND user_id = ? AND status = 0 LIMIT 1"
	err := m.QueryRowNoCacheCtx(ctx, &request, query, groupId, userId)
	if err != nil {
		return nil, err
	}
	return &request, nil
}

// FindByGroupAndStatus 按群组和状态查询申请列表
func (m *customImGroupJoinRequestModel) FindByGroupAndStatus(ctx context.Context, groupId string, status int64, page, pageSize int64) ([]*ImGroupJoinRequest, error) {
	var requests []*ImGroupJoinRequest
	query := "SELECT * FROM im_group_join_request WHERE group_id = ? AND status = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	err := m.QueryRowsNoCacheCtx(ctx, &requests, query, groupId, status, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	return requests, nil
}

// FindByGroup 按群组查询所有申请
func (m *customImGroupJoinRequestModel) FindByGroup(ctx context.Context, groupId string, page, pageSize int64) ([]*ImGroupJoinRequest, error) {
	var requests []*ImGroupJoinRequest
	query := "SELECT * FROM im_group_join_request WHERE group_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	err := m.QueryRowsNoCacheCtx(ctx, &requests, query, groupId, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	return requests, nil
}

// FindByUserId 查询用户发出的申请
func (m *customImGroupJoinRequestModel) FindByUserId(ctx context.Context, userId int64, page, pageSize int64) ([]*ImGroupJoinRequest, error) {
	var requests []*ImGroupJoinRequest
	query := "SELECT * FROM im_group_join_request WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	err := m.QueryRowsNoCacheCtx(ctx, &requests, query, userId, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	return requests, nil
}

// CountPendingByGroup 统计群待处理申请数
func (m *customImGroupJoinRequestModel) CountPendingByGroup(ctx context.Context, groupId string) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM im_group_join_request WHERE group_id = ? AND status = 0"
	err := m.QueryRowNoCacheCtx(ctx, &count, query, groupId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountByUserId 统计用户发出的申请数
func (m *customImGroupJoinRequestModel) CountByUserId(ctx context.Context, userId int64) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM im_group_join_request WHERE user_id = ?"
	err := m.QueryRowNoCacheCtx(ctx, &count, query, userId)
	if err != nil {
		return 0, err
	}
	return count, nil
}
