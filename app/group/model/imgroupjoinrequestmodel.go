package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImGroupJoinRequestModel = (*customImGroupJoinRequestModel)(nil)

type (
	// ImGroupJoinRequestModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImGroupJoinRequestModel.
	ImGroupJoinRequestModel interface {
		imGroupJoinRequestModel
		// 按群组ID和状态查询申请列表(分页)
		FindByGroupIdAndStatus(ctx context.Context, groupId string, status int64, page, pageSize int64) ([]*ImGroupJoinRequest, error)
		// 按用户ID查询申请列表(分页)
		FindByUserId(ctx context.Context, userId uint64, page, pageSize int64) ([]*ImGroupJoinRequest, error)
		// 检查用户是否有待处理的申请
		FindPendingByGroupAndUser(ctx context.Context, groupId string, userId uint64) (*ImGroupJoinRequest, error)
		// 查询用户对某群的最新申请记录（任何状态）
		FindLatestByGroupAndUser(ctx context.Context, groupId string, userId uint64) (*ImGroupJoinRequest, error)
		// 统计群组的申请数量
		CountByGroupIdAndStatus(ctx context.Context, groupId string, status int64) (int64, error)
		// 统计用户的申请数量
		CountByUserId(ctx context.Context, userId uint64) (int64, error)
		// 按多个群组ID和状态查询申请列表(分页)
		FindByGroupIdsAndStatus(ctx context.Context, groupIds []string, status int64, page, pageSize int64) ([]*ImGroupJoinRequest, int64, error)
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

// FindByGroupIdAndStatus 按群组ID和状态查询申请列表(分页)
func (m *customImGroupJoinRequestModel) FindByGroupIdAndStatus(ctx context.Context, groupId string, status int64, page, pageSize int64) ([]*ImGroupJoinRequest, error) {
	var resp []*ImGroupJoinRequest
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where `group_id` = ? and `status` = ? order by `created_at` desc limit ? offset ?", imGroupJoinRequestRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, groupId, status, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// FindByUserId 按用户ID查询申请列表(分页)
func (m *customImGroupJoinRequestModel) FindByUserId(ctx context.Context, userId uint64, page, pageSize int64) ([]*ImGroupJoinRequest, error) {
	var resp []*ImGroupJoinRequest
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where `user_id` = ? order by `created_at` desc limit ? offset ?", imGroupJoinRequestRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, userId, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// FindPendingByGroupAndUser 检查用户是否有待处理的申请
func (m *customImGroupJoinRequestModel) FindPendingByGroupAndUser(ctx context.Context, groupId string, userId uint64) (*ImGroupJoinRequest, error) {
	// 使用生成的方法，status=0表示待处理
	return m.FindOneByGroupIdUserIdStatus(ctx, groupId, userId, 0)
}

// CountByGroupIdAndStatus 统计群组的申请数量
func (m *customImGroupJoinRequestModel) CountByGroupIdAndStatus(ctx context.Context, groupId string, status int64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where `group_id` = ? and `status` = ?", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, groupId, status)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountByUserId 统计用户的申请数量
func (m *customImGroupJoinRequestModel) CountByUserId(ctx context.Context, userId uint64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where `user_id` = ?", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, userId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// FindLatestByGroupAndUser 查询用户对某群的最新申请记录（任何状态）
func (m *customImGroupJoinRequestModel) FindLatestByGroupAndUser(ctx context.Context, groupId string, userId uint64) (*ImGroupJoinRequest, error) {
	var resp ImGroupJoinRequest
	query := fmt.Sprintf("select %s from %s where `group_id` = ? and `user_id` = ? order by `created_at` desc limit 1", imGroupJoinRequestRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, groupId, userId)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByGroupIdsAndStatus 按多个群组ID和状态查询申请列表(分页)
func (m *customImGroupJoinRequestModel) FindByGroupIdsAndStatus(ctx context.Context, groupIds []string, status int64, page, pageSize int64) ([]*ImGroupJoinRequest, int64, error) {
	// 如果群组列表为空，直接返回
	if len(groupIds) == 0 {
		return []*ImGroupJoinRequest{}, 0, nil
	}

	// 构建IN子句的占位符
	placeholders := make([]string, len(groupIds))
	args := make([]interface{}, 0, len(groupIds)+3)
	for i := range groupIds {
		placeholders[i] = "?"
		args = append(args, groupIds[i])
	}
	inClause := fmt.Sprintf("(%s)", strings.Join(placeholders, ","))

	// 先查询总数
	var total int64
	countQuery := fmt.Sprintf("select count(*) from %s where `group_id` in %s and `status` = ?", m.table, inClause)
	countArgs := append(args, status)
	err := m.QueryRowNoCacheCtx(ctx, &total, countQuery, countArgs...)
	if err != nil {
		return nil, 0, err
	}

	// 查询分页数据
	var resp []*ImGroupJoinRequest
	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf("select %s from %s where `group_id` in %s and `status` = ? order by `created_at` desc limit ? offset ?", imGroupJoinRequestRows, m.table, inClause)
	dataArgs := append(args, status, pageSize, offset)
	err = m.QueryRowsNoCacheCtx(ctx, &resp, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}

	return resp, total, nil
}
