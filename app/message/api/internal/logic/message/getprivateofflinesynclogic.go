// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package message

import (
	"context"
	"sort"

	"SkyeIM/app/message/api/internal/svc"
	"SkyeIM/app/message/api/internal/types"
	"SkyeIM/app/message/rpc/message"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPrivateOfflineSyncLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 私聊离线同步（拉取剩余离线消息）
func NewGetPrivateOfflineSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPrivateOfflineSyncLogic {
	return &GetPrivateOfflineSyncLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPrivateOfflineSyncLogic) GetPrivateOfflineSync(req *types.GetPrivateOfflineSyncReq) (resp *types.GetPrivateOfflineSyncResp, err error) {
	// 从JWT中获取当前用户ID
	userId, err := getUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	// 1. 直接获取所有离线消息（优化：通过 RPC 批量拉取）
	// PeerId=0 表示获取用户的 *所有* 未读消息
	unreadResp, err := l.svcCtx.MessageRpc.GetUnreadMessages(l.ctx, &message.GetUnreadMessagesReq{
		UserId: userId,
		PeerId: 0,
	})
	if err != nil {
		logx.Errorf("[GetPrivateOfflineSync] Failed to get unread messages: %v", err)
		return nil, err
	}

	var allMessages []*message.MessageInfo
	if len(unreadResp.List) > 0 {
		allMessages = unreadResp.List
	}

	// 3. 按时间排序（从旧到新）
	// 注意：RPC返回的消息已经是按时间排序的，但这里为了保险再排一次
	// 如果消息来自多个好友，可能需要重新排序
	sortMessagesByTime(allMessages)

	// 4. 分页处理
	totalCount := int64(len(allMessages))
	skip := int(req.Skip)
	limit := int(req.Limit)

	// 边界检查
	if skip < 0 {
		skip = 0
	}
	if limit <= 0 || limit > 200 {
		limit = 100 // 默认100条，最多200条
	}

	// 计算分页
	start := skip
	end := skip + limit

	if start >= len(allMessages) {
		// 没有更多数据
		return &types.GetPrivateOfflineSyncResp{
			List:    []types.MessageInfo{},
			HasMore: false,
			Total:   totalCount,
		}, nil
	}

	if end > len(allMessages) {
		end = len(allMessages)
	}

	pageMessages := allMessages[start:end]

	// 5. 转换为API类型
	list := make([]types.MessageInfo, 0, len(pageMessages))
	for _, msg := range pageMessages {
		list = append(list, types.MessageInfo{
			Id:          msg.Id,
			MsgId:       msg.MsgId,
			FromUserId:  msg.FromUserId,
			ToUserId:    msg.ToUserId,
			ChatType:    msg.ChatType,
			GroupId:     msg.GroupId,
			Content:     msg.Content,
			ContentType: msg.ContentType,
			Status:      msg.Status,
			CreatedAt:   msg.CreatedAt,
			Seq:         msg.Seq,
			AtUserIds:   msg.AtUserIds,
		})
	}

	return &types.GetPrivateOfflineSyncResp{
		List:    list,
		HasMore: end < len(allMessages),
		Total:   totalCount,
	}, nil
}

// sortMessagesByTime 按创建时间排序（从旧到新）
// sortMessagesByTime 按创建时间排序（从旧到新）
func sortMessagesByTime(messages []*message.MessageInfo) {
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt < messages[j].CreatedAt
	})
}
