package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImFriendRequestModel = (*customImFriendRequestModel)(nil)

type (
	// ImFriendRequestModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImFriendRequestModel.
	ImFriendRequestModel interface {
		imFriendRequestModel
		// 自定义方法
		FindByToUserId(ctx context.Context, toUserId uint64, page, pageSize int64) ([]*ImFriendRequest, error)
		CountByToUserId(ctx context.Context, toUserId uint64) (int64, error)
		FindByFromUserId(ctx context.Context, fromUserId uint64, page, pageSize int64) ([]*ImFriendRequest, error)
		CountByFromUserId(ctx context.Context, fromUserId uint64) (int64, error)
		FindPendingRequest(ctx context.Context, fromUserId, toUserId uint64) (*ImFriendRequest, error)
		UpdateStatus(ctx context.Context, id uint64, status int64) error
	}

	customImFriendRequestModel struct {
		*defaultImFriendRequestModel
	}
)

// NewImFriendRequestModel returns a model for the database table.
func NewImFriendRequestModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ImFriendRequestModel {
	return &customImFriendRequestModel{
		defaultImFriendRequestModel: newImFriendRequestModel(conn, c, opts...),
	}
}

// FindByToUserId 获取收到的好友申请列表
func (m *customImFriendRequestModel) FindByToUserId(ctx context.Context, toUserId uint64, page, pageSize int64) ([]*ImFriendRequest, error) {
	var resp []*ImFriendRequest
	query := fmt.Sprintf("select %s from %s where `to_user_id` = ? order by `created_at` desc limit ?, ?", imFriendRequestRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, toUserId, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CountByToUserId 统计收到的好友申请数量
func (m *customImFriendRequestModel) CountByToUserId(ctx context.Context, toUserId uint64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where `to_user_id` = ?", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, toUserId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// FindByFromUserId 获取发出的好友申请列表
func (m *customImFriendRequestModel) FindByFromUserId(ctx context.Context, fromUserId uint64, page, pageSize int64) ([]*ImFriendRequest, error) {
	var resp []*ImFriendRequest
	query := fmt.Sprintf("select %s from %s where `from_user_id` = ? order by `created_at` desc limit ?, ?", imFriendRequestRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, fromUserId, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CountByFromUserId 统计发出的好友申请数量
func (m *customImFriendRequestModel) CountByFromUserId(ctx context.Context, fromUserId uint64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where `from_user_id` = ?", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, fromUserId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// FindPendingRequest 查找待处理的好友申请（防止重复申请）
func (m *customImFriendRequestModel) FindPendingRequest(ctx context.Context, fromUserId, toUserId uint64) (*ImFriendRequest, error) {
	var resp ImFriendRequest
	query := fmt.Sprintf("select %s from %s where `from_user_id` = ? and `to_user_id` = ? and `status` = 0 limit 1", imFriendRequestRows, m.table)
	err := m.QueryRowNoCacheCtx(ctx, &resp, query, fromUserId, toUserId)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateStatus 更新申请状态
func (m *customImFriendRequestModel) UpdateStatus(ctx context.Context, id uint64, status int64) error {
	query := fmt.Sprintf("update %s set `status` = ? where `id` = ?", m.table)
	_, err := m.ExecNoCacheCtx(ctx, query, status, id)
	return err
}
