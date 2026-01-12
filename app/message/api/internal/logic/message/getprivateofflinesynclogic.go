// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package message

import (
	"context"

	"SkyeIM/app/friend/rpc/friend"
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
	userId := l.ctx.Value("userId").(int64)

	// 1. 获取用户的好友列表
	friendResp, err := l.svcCtx.FriendRpc.GetFriendList(l.ctx, &friend.GetFriendListReq{
		UserId:   userId,
		Page:     1,
		PageSize: 10000, // 获取所有好友
	})
	if err != nil {
		logx.Errorf("[GetPrivateOfflineSync] Failed to get friend list: %v", err)
		return nil, err
	}

	// 2. 收集所有好友的未读消息
	var allMessages []*message.MessageInfo
	for _, friendInfo := range friendResp.List {
		// 跳过被拉黑的好友
		if friendInfo.Status == 2 {
			continue
		}

		// 获取与该好友的未读消息
		unreadResp, err := l.svcCtx.MessageRpc.GetUnreadMessages(l.ctx, &message.GetUnreadMessagesReq{
			UserId: userId,
			PeerId: friendInfo.FriendId,
		})
		if err != nil {
			logx.Errorf("[GetPrivateOfflineSync] Failed to get unread from user %d: %v", friendInfo.FriendId, err)
			continue
		}

		if len(unreadResp.List) > 0 {
			allMessages = append(allMessages, unreadResp.List...)
		}
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
		})
	}

	return &types.GetPrivateOfflineSyncResp{
		List:    list,
		HasMore: end < len(allMessages),
		Total:   totalCount,
	}, nil
}

// sortMessagesByTime 按创建时间排序（从旧到新）
func sortMessagesByTime(messages []*message.MessageInfo) {
	// 简单的冒泡排序，生产环境可用sort.Slice
	for i := 0; i < len(messages)-1; i++ {
		for j := 0; j < len(messages)-i-1; j++ {
			if messages[j].CreatedAt > messages[j+1].CreatedAt {
				messages[j], messages[j+1] = messages[j+1], messages[j]
			}
		}
	}
}
