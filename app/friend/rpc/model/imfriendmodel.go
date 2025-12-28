package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImFriendModel = (*customImFriendModel)(nil)

type (
	// ImFriendModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImFriendModel.
	ImFriendModel interface {
		imFriendModel
		// 自定义方法
		FindByUserId(ctx context.Context, userId uint64, status int64, page, pageSize int64) ([]*ImFriend, error)
		CountByUserId(ctx context.Context, userId uint64, status int64) (int64, error)
		FindBlacklist(ctx context.Context, userId uint64, page, pageSize int64) ([]*ImFriend, error)
		CountBlacklist(ctx context.Context, userId uint64) (int64, error)
		DeleteByUserIdFriendId(ctx context.Context, userId, friendId uint64) error
	}

	customImFriendModel struct {
		*defaultImFriendModel
	}
)

// NewImFriendModel returns a model for the database table.
func NewImFriendModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ImFriendModel {
	return &customImFriendModel{
		defaultImFriendModel: newImFriendModel(conn, c, opts...),
	}
}

// FindByUserId 根据用户ID查询好友列表
func (m *customImFriendModel) FindByUserId(ctx context.Context, userId uint64, status int64, page, pageSize int64) ([]*ImFriend, error) {
	var resp []*ImFriend
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `status` = ? order by `created_at` desc limit ?, ?", imFriendRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, userId, status, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CountByUserId 统计用户好友数量
func (m *customImFriendModel) CountByUserId(ctx context.Context, userId uint64, status int64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where `user_id` = ? and `status` = ?", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, userId, status)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// FindBlacklist 获取黑名单列表
func (m *customImFriendModel) FindBlacklist(ctx context.Context, userId uint64, page, pageSize int64) ([]*ImFriend, error) {
	var resp []*ImFriend
	query := fmt.Sprintf("select %s from %s where `user_id` = ? and `status` = 2 order by `updated_at` desc limit ?, ?", imFriendRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, userId, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CountBlacklist 统计黑名单数量
func (m *customImFriendModel) CountBlacklist(ctx context.Context, userId uint64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where `user_id` = ? and `status` = 2", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, userId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteByUserIdFriendId 根据用户ID和好友ID删除好友关系
func (m *customImFriendModel) DeleteByUserIdFriendId(ctx context.Context, userId, friendId uint64) error {
	// 先查找记录以便清除缓存
	data, err := m.FindOneByUserIdFriendId(ctx, userId, friendId)
	if err != nil {
		if err == ErrNotFound {
			return nil
		}
		return err
	}

	imFriendIdKey := fmt.Sprintf("%s%v", cacheImFriendIdPrefix, data.Id)
	imFriendUserIdFriendIdKey := fmt.Sprintf("%s%v:%v", cacheImFriendUserIdFriendIdPrefix, userId, friendId)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `user_id` = ? and `friend_id` = ?", m.table)
		return conn.ExecCtx(ctx, query, userId, friendId)
	}, imFriendIdKey, imFriendUserIdFriendIdKey)
	return err
}
