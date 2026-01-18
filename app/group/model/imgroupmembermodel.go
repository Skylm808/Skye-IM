package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImGroupMemberModel = (*customImGroupMemberModel)(nil)

type (
	// ImGroupMemberModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImGroupMemberModel.
	ImGroupMemberModel interface {
		imGroupMemberModel
		FindByGroupId(ctx context.Context, groupId string) ([]*ImGroupMember, error)
		FindGroupsByUserId(ctx context.Context, userId int64, page, pageSize int32) ([]*ImGroupMember, error)
		CountByGroupId(ctx context.Context, groupId string) (int64, error)
		CountByUserId(ctx context.Context, userId int64) (int64, error)
		DeleteByGroupIdUserId(ctx context.Context, groupId string, userId int64) error
		UpdateReadSeq(ctx context.Context, groupId string, userId int64, readSeq uint64) error
		FindManagedGroupsByUserId(ctx context.Context, userId int64) ([]string, error)
	}

	customImGroupMemberModel struct {
		*defaultImGroupMemberModel
	}
)

// NewImGroupMemberModel returns a model for the database table.
func NewImGroupMemberModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ImGroupMemberModel {
	return &customImGroupMemberModel{
		defaultImGroupMemberModel: newImGroupMemberModel(conn, c, opts...),
	}
}

// FindByGroupId 查询群组所有成员
func (m *customImGroupMemberModel) FindByGroupId(ctx context.Context, groupId string) ([]*ImGroupMember, error) {
	var resp []*ImGroupMember
	query := fmt.Sprintf("select %s from %s where `group_id` = ?", imGroupMemberRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, groupId)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

// FindGroupsByUserId 查询用户加入的所有群组（分页）
func (m *customImGroupMemberModel) FindGroupsByUserId(ctx context.Context, userId int64, page, pageSize int32) ([]*ImGroupMember, error) {
	var resp []*ImGroupMember
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("select %s from %s where `user_id` = ? order by `joined_at` desc limit ?,?", imGroupMemberRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, userId, offset, pageSize)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

// CountByGroupId 统计群组成员数
func (m *customImGroupMemberModel) CountByGroupId(ctx context.Context, groupId string) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where `group_id` = ?", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, groupId)
	return count, err
}

// CountByUserId 统计用户加入的群组数
func (m *customImGroupMemberModel) CountByUserId(ctx context.Context, userId int64) (int64, error) {
	var count int64
	query := fmt.Sprintf("select count(*) from %s where `user_id` = ?", m.table)
	err := m.QueryRowNoCacheCtx(ctx, &count, query, userId)
	return count, err
}

// DeleteByGroupIdUserId 根据群组ID和用户ID删除成员
func (m *customImGroupMemberModel) DeleteByGroupIdUserId(ctx context.Context, groupId string, userId int64) error {
	// 先查询获取缓存key
	data, err := m.FindOneByGroupIdUserId(ctx, groupId, userId)
	if err != nil {
		return err
	}

	imGroupMemberGroupIdUserIdKey := fmt.Sprintf("%s%v:%v", cacheImGroupMemberGroupIdUserIdPrefix, groupId, userId)
	imGroupMemberIdKey := fmt.Sprintf("%s%v", cacheImGroupMemberIdPrefix, data.Id)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `group_id` = ? and `user_id` = ?", m.table)
		return conn.ExecCtx(ctx, query, groupId, userId)
	}, imGroupMemberGroupIdUserIdKey, imGroupMemberIdKey)
	return err
}

// UpdateReadSeq 更新用户在群组的已读Seq
func (m *customImGroupMemberModel) UpdateReadSeq(ctx context.Context, groupId string, userId int64, readSeq uint64) error {
	// 先查询拿到 ID，用于精准清理缓存；更新使用 SQL 保证 read_seq 单调递增。
	data, err := m.FindOneByGroupIdUserId(ctx, groupId, userId)
	if err != nil {
		return err
	}

	imGroupMemberGroupIdUserIdKey := fmt.Sprintf("%s%v:%v", cacheImGroupMemberGroupIdUserIdPrefix, groupId, userId)
	imGroupMemberIdKey := fmt.Sprintf("%s%v", cacheImGroupMemberIdPrefix, data.Id)

	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		// GREATEST 保证 read_seq 不会回退（多端并发上报时取最大值）
		query := fmt.Sprintf("update %s set `read_seq` = GREATEST(`read_seq`, ?) where `group_id` = ? and `user_id` = ?", m.table)
		return conn.ExecCtx(ctx, query, readSeq, groupId, userId)
	}, imGroupMemberGroupIdUserIdKey, imGroupMemberIdKey)
	return err
}

// FindManagedGroupsByUserId 查询用户作为管理员/群主的所有群组ID
func (m *customImGroupMemberModel) FindManagedGroupsByUserId(ctx context.Context, userId int64) ([]string, error) {
	var groupIds []string
	query := fmt.Sprintf("select `group_id` from %s where `user_id` = ? and `role` in (1, 2)", m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &groupIds, query, userId)
	switch err {
	case nil:
		return groupIds, nil
	default:
		return nil, err
	}
}
