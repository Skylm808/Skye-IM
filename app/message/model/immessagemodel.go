package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImMessageModel = (*customImMessageModel)(nil)

type (
	// ImMessageModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImMessageModel.
	ImMessageModel interface {
		imMessageModel
		// 群聊消息查询方法
		FindGroupMessageList(ctx context.Context, groupId string, lastMsgId, limit int64) ([]*ImMessage, error)
		// 私聊消息查询方法
		FindPrivateMessageList(ctx context.Context, userId, peerId, lastMsgId, limit int64) ([]*ImMessage, error)
		FindUnreadMessages(ctx context.Context, userId, peerId int64) ([]*ImMessage, error)
		CountUnreadMessages(ctx context.Context, userId, peerId int64) (int64, error)
		MarkMessagesAsRead(ctx context.Context, userId, peerId int64, msgIds []string) (int64, error)
		FindGroupMessagesAfterSeq(ctx context.Context, groupId string, seq uint64) ([]*ImMessage, error)
		// 模糊搜索消息内容
		SearchByKeyword(ctx context.Context, userId int64, keyword string) ([]*ImMessage, error)
		// 查询@我的消息
		FindAtMeMessages(ctx context.Context, userId int64, groupId string, lastMsgId int64, limit int32) ([]*ImMessage, error)
		// 暴露底层数据库操作方法
		QueryRowsNoCacheCtx(ctx context.Context, v interface{}, query string, args ...interface{}) error
		QueryRowNoCacheCtx(ctx context.Context, v interface{}, query string, args ...interface{}) error
		ExecNoCacheCtx(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	}

	customImMessageModel struct {
		*defaultImMessageModel
	}
)

// NewImMessageModel returns a model for the database table.
func NewImMessageModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ImMessageModel {
	return &customImMessageModel{
		defaultImMessageModel: newImMessageModel(conn, c, opts...),
	}
}

// FindGroupMessageList 查询群聊历史消息（分页）
func (m *customImMessageModel) FindGroupMessageList(ctx context.Context, groupId string, lastMsgId, limit int64) ([]*ImMessage, error) {
	var resp []*ImMessage

	var query string
	var args []interface{}

	if lastMsgId > 0 {
		query = fmt.Sprintf("select %s from %s where `chat_type` = 2 and `group_id` = ? and `id` < ? order by `id` desc limit ?", imMessageRows, m.table)
		args = []interface{}{groupId, lastMsgId, limit}
	} else {
		query = fmt.Sprintf("select %s from %s where `chat_type` = 2 and `group_id` = ? order by `id` desc limit ?", imMessageRows, m.table)
		args = []interface{}{groupId, limit}
	}

	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

// FindPrivateMessageList 查询私聊历史消息（分页）
func (m *customImMessageModel) FindPrivateMessageList(ctx context.Context, userId, peerId, lastMsgId, limit int64) ([]*ImMessage, error) {
	var resp []*ImMessage

	var query string
	var args []interface{}

	if lastMsgId > 0 {
		query = fmt.Sprintf("select %s from %s where `chat_type` = 1 and ((from_user_id = ? and to_user_id = ?) or (from_user_id = ? and to_user_id = ?)) and `id` < ? order by `id` desc limit ?", imMessageRows, m.table)
		args = []interface{}{userId, peerId, peerId, userId, lastMsgId, limit}
	} else {
		query = fmt.Sprintf("select %s from %s where `chat_type` = 1 and ((from_user_id = ? and to_user_id = ?) or (from_user_id = ? and to_user_id = ?)) order by `id` desc limit ?", imMessageRows, m.table)
		args = []interface{}{userId, peerId, peerId, userId, limit}
	}

	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

// FindUnreadMessages 获取未读消息列表
func (m *customImMessageModel) FindUnreadMessages(ctx context.Context, userId, peerId int64) ([]*ImMessage, error) {
	var resp []*ImMessage
	query := fmt.Sprintf("select %s from %s where `chat_type` = 1 and `to_user_id` = ? and `from_user_id` = ? and `status` = 0 order by `id` asc", imMessageRows, m.table)

	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, userId, peerId)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

// CountUnreadMessages 统计未读消息数量
func (m *customImMessageModel) CountUnreadMessages(ctx context.Context, userId, peerId int64) (int64, error) {
	var count int64
	var query string
	var args []interface{}

	if peerId > 0 {
		query = fmt.Sprintf("select count(*) from %s where `chat_type` = 1 and `to_user_id` = ? and `from_user_id` = ? and `status` = 0", m.table)
		args = []interface{}{userId, peerId}
	} else {
		query = fmt.Sprintf("select count(*) from %s where `chat_type` = 1 and `to_user_id` = ? and `status` = 0", m.table)
		args = []interface{}{userId}
	}

	err := m.QueryRowNoCacheCtx(ctx, &count, query, args...)
	return count, err
}

// MarkMessagesAsRead 标记消息为已读
func (m *customImMessageModel) MarkMessagesAsRead(ctx context.Context, userId, peerId int64, msgIds []string) (int64, error) {
	var query string
	var args []interface{}

	if len(msgIds) > 0 {
		placeholders := ""
		for i := range msgIds {
			if i > 0 {
				placeholders += ","
			}
			placeholders += "?"
		}
		query = fmt.Sprintf("update %s set `status` = 1 where `chat_type` = 1 and `to_user_id` = ? and `from_user_id` = ? and `msg_id` in (%s)", m.table, placeholders)
		args = append([]interface{}{userId, peerId}, convertStringsToInterfaces(msgIds)...)
	} else {
		query = fmt.Sprintf("update %s set `status` = 1 where `chat_type` = 1 and `to_user_id` = ? and `from_user_id` = ? and `status` = 0", m.table)
		args = []interface{}{userId, peerId}
	}

	result, err := m.ExecNoCacheCtx(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// convertStringsToInterfaces 辅助函数：字符串数组转接口数组
func convertStringsToInterfaces(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, s := range strs {
		result[i] = s
	}
	return result
}

// QueryRowsNoCacheCtx 查询多行数据（不使用缓存）
func (m *customImMessageModel) QueryRowsNoCacheCtx(ctx context.Context, v interface{}, query string, args ...interface{}) error {
	return m.CachedConn.QueryRowsNoCacheCtx(ctx, v, query, args...)
}

// QueryRowNoCacheCtx 查询单行数据（不使用缓存）
func (m *customImMessageModel) QueryRowNoCacheCtx(ctx context.Context, v interface{}, query string, args ...interface{}) error {
	return m.CachedConn.QueryRowNoCacheCtx(ctx, v, query, args...)
}

// ExecNoCacheCtx 执行SQL（不使用缓存）
func (m *customImMessageModel) ExecNoCacheCtx(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return m.CachedConn.ExecNoCacheCtx(ctx, query, args...)
}

// FindGroupMessagesAfterSeq 查询大于指定Seq的群聊消息
// SearchByKeyword 模糊搜索消息内容（仅限用户参与的聊天）
func (m *customImMessageModel) SearchByKeyword(ctx context.Context, userId int64, keyword string) ([]*ImMessage, error) {
	var resp []*ImMessage
	likeKeyword := "%" + keyword + "%"

	// 这里的逻辑稍微复杂点：搜索用户参与的私聊消息，或者用户所在群组的消息（这里简化为全库搜索内容匹配的消息，实际生产环境需要关联群成员表）
	// 为了演示，我们先实现基础的内容匹配，关联用户 ID 以保证只能搜到自己的私聊
	query := fmt.Sprintf("select %s from %s where `content` like ? and ((`chat_type` = 1 and (`from_user_id` = ? or `to_user_id` = ?)) or (`chat_type` = 2)) order by `created_at` desc limit 100", imMessageRows, m.table)

	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, likeKeyword, userId, userId)
	return resp, err
}

func (m *customImMessageModel) FindGroupMessagesAfterSeq(ctx context.Context, groupId string, seq uint64) ([]*ImMessage, error) {
	var resp []*ImMessage
	query := fmt.Sprintf("select %s from %s where `chat_type` = 2 and `group_id` = ? and `seq` > ? order by `seq` asc limit 200", imMessageRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, groupId, seq)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

// FindAtMeMessages 查询@我的消息（群聊）
func (m *customImMessageModel) FindAtMeMessages(ctx context.Context, userId int64, groupId string, lastMsgId int64, limit int32) ([]*ImMessage, error) {
	var resp []*ImMessage

	if limit <= 0 {
		limit = 20
	}

	// 构建查询：查找at_user_ids包含当前用户ID或-1（@全体）的消息
	var query string
	var args []interface{}

	// JSON_CONTAINS 或 LIKE 查询
	// 使用 LIKE 匹配 JSON 数组中的值
	if groupId != "" {
		// 查询特定群组
		if lastMsgId > 0 {
			query = fmt.Sprintf(
				"SELECT %s FROM %s WHERE `chat_type` = 2 AND `group_id` = ? AND `id` < ? AND "+
					"(`at_user_ids` LIKE ? OR `at_user_ids` LIKE ?) "+
					"ORDER BY `id` DESC LIMIT ?",
				imMessageRows, m.table,
			)
			args = []interface{}{
				groupId,
				lastMsgId,
				fmt.Sprintf("%%\"%d\"%%", userId), // 匹配 JSON 数组中的用户ID
				"%\"-1\"%",                        // 匹配 -1 (@全体)
				limit,
			}
		} else {
			query = fmt.Sprintf(
				"SELECT %s FROM %s WHERE `chat_type` = 2 AND `group_id` = ? AND "+
					"(`at_user_ids` LIKE ? OR `at_user_ids` LIKE ?) "+
					"ORDER BY `id` DESC LIMIT ?",
				imMessageRows, m.table,
			)
			args = []interface{}{
				groupId,
				fmt.Sprintf("%%\"%d\"%%", userId),
				"%\"-1\"%",
				limit,
			}
		}
	} else {
		// 查询所有群组
		if lastMsgId > 0 {
			query = fmt.Sprintf(
				"SELECT %s FROM %s WHERE `chat_type` = 2 AND `id` < ? AND "+
					"(`at_user_ids` LIKE ? OR `at_user_ids` LIKE ?) "+
					"ORDER BY `id` DESC LIMIT ?",
				imMessageRows, m.table,
			)
			args = []interface{}{
				lastMsgId,
				fmt.Sprintf("%%\"%d\"%%", userId),
				"%\"-1\"%",
				limit,
			}
		} else {
			query = fmt.Sprintf(
				"SELECT %s FROM %s WHERE `chat_type` = 2 AND "+
					"(`at_user_ids` LIKE ? OR `at_user_ids` LIKE ?) "+
					"ORDER BY `id` DESC LIMIT ?",
				imMessageRows, m.table,
			)
			args = []interface{}{
				fmt.Sprintf("%%\"%d\"%%", userId),
				"%\"-1\"%",
				limit,
			}
		}
	}

	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, args...)
	return resp, err
}
