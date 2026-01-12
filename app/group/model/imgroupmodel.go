package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImGroupModel = (*customImGroupModel)(nil)

type (
	// ImGroupModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImGroupModel.
	ImGroupModel interface {
		imGroupModel
		SearchByKeyword(ctx context.Context, keyword string) ([]*ImGroup, error)
		FindOneByName(ctx context.Context, name string) (*ImGroup, error)
	}

	customImGroupModel struct {
		*defaultImGroupModel
	}
)

func (m *customImGroupModel) SearchByKeyword(ctx context.Context, keyword string) ([]*ImGroup, error) {
	var resp []*ImGroup
	likeKeyword := "%" + keyword + "%"
	// 同时模糊匹配群号和群名
	query := fmt.Sprintf("select %s from %s where `group_id` like ? or `name` like ? limit 50", imGroupRows, m.table)
	err := m.CachedConn.QueryRowsNoCacheCtx(ctx, &resp, query, likeKeyword, likeKeyword)
	return resp, err
}

func (m *customImGroupModel) FindOneByName(ctx context.Context, name string) (*ImGroup, error) {
	var resp ImGroup
	query := fmt.Sprintf("select %s from %s where `name` = ? limit 1", imGroupRows, m.table)
	err := m.CachedConn.QueryRowNoCacheCtx(ctx, &resp, query, name)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// NewImGroupModel returns a model for the database table.
func NewImGroupModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ImGroupModel {
	return &customImGroupModel{
		defaultImGroupModel: newImGroupModel(conn, c, opts...),
	}
}
