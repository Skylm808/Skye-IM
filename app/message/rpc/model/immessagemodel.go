package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImMessageModel = (*customImMessageModel)(nil)

type (
	// ImMessageModel interface with custom methods
	ImMessageModel interface {
		Insert(ctx context.Context, data *ImMessage) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*ImMessage, error)
		FindOneByMsgId(ctx context.Context, msgId string) (*ImMessage, error)
		Update(ctx context.Context, data *ImMessage) error
		Delete(ctx context.Context, id int64) error
		// 自定义方法
		GetMessageList(ctx context.Context, userId, peerId, lastMsgId int64, limit int) ([]*ImMessage, error)
		GetUnreadMessages(ctx context.Context, userId, peerId int64) ([]*ImMessage, error)
		GetUnreadCount(ctx context.Context, userId, peerId int64) (int64, error)
		MarkAsRead(ctx context.Context, userId, peerId int64, msgIds []string) (int64, error)
	}

	// ImMessage 消息实体
	ImMessage struct {
		Id          int64     `db:"id"`
		MsgId       string    `db:"msg_id"`
		FromUserId  int64     `db:"from_user_id"`
		ToUserId    int64     `db:"to_user_id"`
		Content     string    `db:"content"`
		ContentType int64     `db:"content_type"`
		Status      int64     `db:"status"`
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
	}

	customImMessageModel struct {
		conn  sqlx.SqlConn
		cache cache.Cache
		table string
	}
)

// NewImMessageModel 创建消息模型
func NewImMessageModel(conn sqlx.SqlConn, c cache.CacheConf) ImMessageModel {
	return &customImMessageModel{
		conn:  conn,
		cache: cache.New(c, nil, cache.NewStat("imMessage"), nil),
		table: "im_message",
	}
}

// Insert 插入消息
func (m *customImMessageModel) Insert(ctx context.Context, data *ImMessage) (sql.Result, error) {
	query := fmt.Sprintf("INSERT INTO %s (msg_id, from_user_id, to_user_id, content, content_type, status) VALUES (?, ?, ?, ?, ?, ?)", m.table)
	return m.conn.ExecCtx(ctx, query, data.MsgId, data.FromUserId, data.ToUserId, data.Content, data.ContentType, data.Status)
}

// FindOne 根据ID查找消息
func (m *customImMessageModel) FindOne(ctx context.Context, id int64) (*ImMessage, error) {
	query := fmt.Sprintf("SELECT id, msg_id, from_user_id, to_user_id, content, content_type, status, created_at, updated_at FROM %s WHERE id = ?", m.table)
	var resp ImMessage
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// FindOneByMsgId 根据MsgId查找消息
func (m *customImMessageModel) FindOneByMsgId(ctx context.Context, msgId string) (*ImMessage, error) {
	query := fmt.Sprintf("SELECT id, msg_id, from_user_id, to_user_id, content, content_type, status, created_at, updated_at FROM %s WHERE msg_id = ?", m.table)
	var resp ImMessage
	err := m.conn.QueryRowCtx(ctx, &resp, query, msgId)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update 更新消息
func (m *customImMessageModel) Update(ctx context.Context, data *ImMessage) error {
	query := fmt.Sprintf("UPDATE %s SET content = ?, content_type = ?, status = ? WHERE id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.Content, data.ContentType, data.Status, data.Id)
	return err
}

// Delete 删除消息
func (m *customImMessageModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// GetMessageList 获取历史消息列表
func (m *customImMessageModel) GetMessageList(ctx context.Context, userId, peerId, lastMsgId int64, limit int) ([]*ImMessage, error) {
	var list []*ImMessage
	var query string
	var err error

	if lastMsgId > 0 {
		// 分页查询：获取 lastMsgId 之前的消息
		query = fmt.Sprintf(`
			SELECT id, msg_id, from_user_id, to_user_id, content, content_type, status, created_at, updated_at 
			FROM %s 
			WHERE ((from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?))
			AND id < ?
			ORDER BY id DESC 
			LIMIT ?`, m.table)
		err = m.conn.QueryRowsCtx(ctx, &list, query, userId, peerId, peerId, userId, lastMsgId, limit)
	} else {
		// 首次查询：获取最新消息
		query = fmt.Sprintf(`
			SELECT id, msg_id, from_user_id, to_user_id, content, content_type, status, created_at, updated_at 
			FROM %s 
			WHERE (from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)
			ORDER BY id DESC 
			LIMIT ?`, m.table)
		err = m.conn.QueryRowsCtx(ctx, &list, query, userId, peerId, peerId, userId, limit)
	}

	if err != nil {
		return nil, err
	}
	return list, nil
}

// GetUnreadMessages 获取未读消息
func (m *customImMessageModel) GetUnreadMessages(ctx context.Context, userId, peerId int64) ([]*ImMessage, error) {
	var list []*ImMessage
	query := fmt.Sprintf(`
		SELECT id, msg_id, from_user_id, to_user_id, content, content_type, status, created_at, updated_at 
		FROM %s 
		WHERE to_user_id = ? AND from_user_id = ? AND status = 0
		ORDER BY id ASC`, m.table)
	err := m.conn.QueryRowsCtx(ctx, &list, query, userId, peerId)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// GetUnreadCount 获取未读消息数量
func (m *customImMessageModel) GetUnreadCount(ctx context.Context, userId, peerId int64) (int64, error) {
	var count int64
	var query string
	var err error

	if peerId > 0 {
		// 获取与特定用户的未读消息数
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE to_user_id = ? AND from_user_id = ? AND status = 0", m.table)
		err = m.conn.QueryRowCtx(ctx, &count, query, userId, peerId)
	} else {
		// 获取所有未读消息数
		query = fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE to_user_id = ? AND status = 0", m.table)
		err = m.conn.QueryRowCtx(ctx, &count, query, userId)
	}

	if err != nil {
		return 0, err
	}
	return count, nil
}

// MarkAsRead 标记消息为已读
func (m *customImMessageModel) MarkAsRead(ctx context.Context, userId, peerId int64, msgIds []string) (int64, error) {
	var result sql.Result
	var err error

	if len(msgIds) > 0 {
		// 标记指定消息为已读
		placeholders := ""
		args := make([]interface{}, 0, len(msgIds)+2)
		args = append(args, userId, peerId)
		for i, msgId := range msgIds {
			if i > 0 {
				placeholders += ", "
			}
			placeholders += "?"
			args = append(args, msgId)
		}
		query := fmt.Sprintf("UPDATE %s SET status = 1 WHERE to_user_id = ? AND from_user_id = ? AND msg_id IN (%s) AND status = 0", m.table, placeholders)
		result, err = m.conn.ExecCtx(ctx, query, args...)
	} else {
		// 标记所有与该用户的未读消息为已读
		query := fmt.Sprintf("UPDATE %s SET status = 1 WHERE to_user_id = ? AND from_user_id = ? AND status = 0", m.table)
		result, err = m.conn.ExecCtx(ctx, query, userId, peerId)
	}

	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
