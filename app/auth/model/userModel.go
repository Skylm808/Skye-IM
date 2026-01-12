package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserModel.
	UserModel interface {
		userModel
		SearchByKeyword(ctx context.Context, keyword string) ([]*User, error)
	}

	customUserModel struct {
		*defaultUserModel
	}
)

func (m *customUserModel) SearchByKeyword(ctx context.Context, keyword string) ([]*User, error) {
	var resp []*User
	likeKeyword := "%" + keyword + "%"
	query := fmt.Sprintf("select %s from %s where `username` like ? or `nickname` like ? or `email` like ? or `phone` like ? limit 50", userRows, m.table)
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, likeKeyword, likeKeyword, likeKeyword, likeKeyword)
	return resp, err
}

// NewUserModel returns a model for the database table.
func NewUserModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) UserModel {
	return &customUserModel{
		defaultUserModel: newUserModel(conn, c, opts...),
	}
}
