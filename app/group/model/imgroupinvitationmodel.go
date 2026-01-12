package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ImGroupInvitationModel = (*customImGroupInvitationModel)(nil)

type (
	// ImGroupInvitationModel is an interface to be customized, add more methods here,
	// and implement the added methods in customImGroupInvitationModel.
	ImGroupInvitationModel interface {
		imGroupInvitationModel
		// 查询被邀请人收到的邀请列表（按状态）
		FindByInviteeIdAndStatus(ctx context.Context, inviteeId int64, status int64, page, pageSize int64) ([]*ImGroupInvitation, error)
		// 查询被邀请人收到的所有邀请
		FindByInviteeId(ctx context.Context, inviteeId int64, page, pageSize int64) ([]*ImGroupInvitation, error)
		// 查询邀请人发出的邀请
		FindByInviterId(ctx context.Context, inviterId int64, page, pageSize int64) ([]*ImGroupInvitation, error)
		// 检查群组和被邀请人是否已有待处理的邀请
		FindPendingByGroupAndInvitee(ctx context.Context, groupId string, inviteeId int64) (*ImGroupInvitation, error)
		// 统计被邀请人的邀请数量
		CountByInviteeId(ctx context.Context, inviteeId int64) (int64, error)
		// 统计邀请人的邀请数量
		CountByInviterId(ctx context.Context, inviterId int64) (int64, error)
	}

	customImGroupInvitationModel struct {
		*defaultImGroupInvitationModel
	}
)

// NewImGroupInvitationModel returns a model for the database table.
func NewImGroupInvitationModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) ImGroupInvitationModel {
	return &customImGroupInvitationModel{
		defaultImGroupInvitationModel: newImGroupInvitationModel(conn, c, opts...),
	}
}

// FindByInviteeIdAndStatus 查询被邀请人收到的邀请列表（按状态）
func (m *customImGroupInvitationModel) FindByInviteeIdAndStatus(ctx context.Context, inviteeId int64, status int64, page, pageSize int64) ([]*ImGroupInvitation, error) {
	var invitations []*ImGroupInvitation
	query := "SELECT * FROM im_group_invitation WHERE invitee_id = ? AND status = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	err := m.QueryRowsNoCacheCtx(ctx, &invitations, query, inviteeId, status, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// FindByInviteeId 查询被邀请人收到的所有邀请
func (m *customImGroupInvitationModel) FindByInviteeId(ctx context.Context, inviteeId int64, page, pageSize int64) ([]*ImGroupInvitation, error) {
	var invitations []*ImGroupInvitation
	query := "SELECT * FROM im_group_invitation WHERE invitee_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	err := m.QueryRowsNoCacheCtx(ctx, &invitations, query, inviteeId, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// FindByInviterId 查询邀请人发出的邀请
func (m *customImGroupInvitationModel) FindByInviterId(ctx context.Context, inviterId int64, page, pageSize int64) ([]*ImGroupInvitation, error) {
	var invitations []*ImGroupInvitation
	query := "SELECT * FROM im_group_invitation WHERE inviter_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?"
	err := m.QueryRowsNoCacheCtx(ctx, &invitations, query, inviterId, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

// FindPendingByGroupAndInvitee 检查群组和被邀请人是否已有待处理的邀请
func (m *customImGroupInvitationModel) FindPendingByGroupAndInvitee(ctx context.Context, groupId string, inviteeId int64) (*ImGroupInvitation, error) {
	var invitation ImGroupInvitation
	query := "SELECT * FROM im_group_invitation WHERE group_id = ? AND invitee_id = ? AND status = 0 LIMIT 1"
	err := m.QueryRowNoCacheCtx(ctx, &invitation, query, groupId, inviteeId)
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

// CountByInviteeId 统计被邀请人的邀请数量
func (m *customImGroupInvitationModel) CountByInviteeId(ctx context.Context, inviteeId int64) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM im_group_invitation WHERE invitee_id = ?"
	err := m.QueryRowNoCacheCtx(ctx, &count, query, inviteeId)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountByInviterId 统计邀请人的邀请数量
func (m *customImGroupInvitationModel) CountByInviterId(ctx context.Context, inviterId int64) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM im_group_invitation WHERE inviter_id = ?"
	err := m.QueryRowNoCacheCtx(ctx, &count, query, inviterId)
	if err != nil {
		return 0, err
	}
	return count, nil
}
