package conn

// Hub - WebSocket 连接管理和消息路由中心
//
// 职责：
// 1. 连接管理：维护所有在线用户的连接映射，处理注册/注销
// 2. 消息路由：将消息路由到指定的一个或多个客户端
//    - 私聊路由：SendToUser() - 直接查表发送（同步，O(1)）
//    - 群聊路由：SendToGroup() - 异步查询成员并批量发送（异步，避免阻塞）
// 3. 状态通知：通知好友上线/下线状态，通知群组事件
//
// 设计说明：
// - 私聊使用同步发送：因为只需要 O(1) 查表，无需异步
// - 群聊使用异步发送：因为需要查询群成员（可能RPC调用），为避免阻塞使用 channel

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"SkyeIM/app/friend/rpc/friend"
	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/ws/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// Hub 维护活跃的客户端连接集合
type Hub struct {
	// 在线用户映射: userId -> Client
	clients map[int64]*Client

	// 注册请求通道
	register chan *Client

	// 注销请求通道
	unregister chan *Client

	// 群组消息通道
	groupMessage chan *GroupMessage

	// 服务上下文（用于调用 RPC）
	svcCtx *svc.ServiceContext

	// 互斥锁
	mu sync.RWMutex
}

// NewHub 创建新的Hub
func NewHub(svcCtx *svc.ServiceContext) *Hub {
	return &Hub{
		clients:      make(map[int64]*Client),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		groupMessage: make(chan *GroupMessage, 256),
		svcCtx:       svcCtx,
	}
}

// Run 启动Hub的消息循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			var toClose *Client
			h.mu.Lock()
			// 如果已有连接，先关闭旧连接
			if oldClient, ok := h.clients[client.UserId]; ok {
				toClose = oldClient
			}
			h.clients[client.UserId] = client
			h.mu.Unlock()
			if toClose != nil {
				toClose.Close()
			}
			logx.Infof("[Hub] User %d connected, total online: %d", client.UserId, len(h.clients))

			// 通知该用户的好友上线
			h.notifyOnlineStatus(client.UserId, true)

		case client := <-h.unregister:
			var toClose *Client
			h.mu.Lock()
			if c, ok := h.clients[client.UserId]; ok && c == client {
				delete(h.clients, client.UserId)
				toClose = client
			}
			h.mu.Unlock()
			if toClose != nil {
				toClose.Close()
			}
			logx.Infof("[Hub] User %d disconnected, total online: %d", client.UserId, len(h.clients))

			// 通知该用户的好友下线
			h.notifyOnlineStatus(client.UserId, false)

		case msg := <-h.groupMessage:
			// ✅ 启动一个新的协程去处理，Hub 主循环瞬间释放，立马可以去处理下一个请求
			// ✅ 即使 routeGroupMessage 卡住 10秒，也不影响别人登录/退出
			go h.routeGroupMessage(msg)
		}
	}
}

// ==================== 连接管理 ====================

// Register 注册客户端（用户上线）
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端（用户下线）
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// IsOnline 检查用户是否在线
func (h *Hub) IsOnline(userId int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userId]
	return ok
}

// GetOnlineUsers 获取在线用户列表
func (h *Hub) GetOnlineUsers() []int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	users := make([]int64, 0, len(h.clients))
	for uid := range h.clients {
		users = append(users, uid)
	}
	return users
}

// OnlineCount 获取在线用户数
func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// ==================== 消息路由 ====================

// SendToUser 路由私聊消息
func (h *Hub) SendToUser(userId int64, msg *Message) bool {
	h.mu.RLock()
	client, ok := h.clients[userId]
	h.mu.RUnlock()

	if !ok {
		return false // 用户不在线
	}

	// 直接发送 Message 对象，WritePump 会负责序列化
	select {
	case client.send <- msg:
		logx.Infof("[Hub] Sent message to user %d, type: %s", userId, msg.Type)
		return true
	default:
		// send channel 满了，说明客户端很慢或已挂，主动关闭连接
		logx.Errorf("[Hub] User %d send buffer full, closing connection", userId)

		// 从 clients map 中移除并关闭连接
		h.mu.Lock()
		delete(h.clients, userId)
		h.mu.Unlock()

		// 关闭连接，让客户端重连后拉取离线消息
		client.Close()

		return false
	}
}

// SendToGroup 路由群聊消息（异步，通过 channel 处理）
// 为什么异步：需要查询群成员列表（可能涉及 RPC 调用），为避免阻塞使用 channel
func (h *Hub) SendToGroup(groupId string, msg *Message, excludeUsers []int64) {
	h.groupMessage <- &GroupMessage{
		GroupId:      groupId,
		Message:      msg,
		ExcludeUsers: excludeUsers,
	}
}

// ==================== 状态通知 ====================

// NotifyGroupEvent 通知群组事件（加入、退出、踢出等）
func (h *Hub) NotifyGroupEvent(groupId string, eventType string, eventData interface{}) {
	data, _ := json.Marshal(eventData)
	msg := &Message{
		Type: eventType,
		Data: json.RawMessage(data),
	}

	// 发送到群组消息通道（通知所有成员）
	h.groupMessage <- &GroupMessage{
		GroupId:      groupId,
		Message:      msg,
		ExcludeUsers: []int64{}, // 不排除任何人，通知所有成员
	}

	logx.Infof("[Hub] Notified group %s event: %s", groupId, eventType)
}

// notifyOnlineStatus 通知好友在线状态变化
func (h *Hub) notifyOnlineStatus(userId int64, online bool) {
	statusType := "offline"
	if online {
		statusType = "online"
	}

	msg := &Message{
		Type: statusType,
		Data: mustMarshalMap(map[string]interface{}{
			"userId":    userId,
			"timestamp": time.Now().Unix(),
		}),
	}

	// 从 Friend RPC 获取好友列表
	ctx := context.Background()
	resp, err := h.svcCtx.FriendRpc.GetFriendList(ctx, &friend.GetFriendListReq{
		UserId:   userId,
		Page:     1,
		PageSize: 10000, // 获取所有好友
	})

	if err != nil {
		logx.Errorf("[Hub] Failed to get friend list for user %d: %v", userId, err)
		return
	}

	// 只通知在线的好友
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, friendInfo := range resp.List {
		// 跳过被拉黑的好友
		if friendInfo.Status == 2 {
			continue
		}

		// 检查好友是否在线
		if client, ok := h.clients[friendInfo.FriendId]; ok {
			// 直接发送 Message 对象
			select {
			case client.send <- msg:
				logx.Infof("[Hub] Notified friend %d about user %d %s", friendInfo.FriendId, userId, statusType)
			default:
				// send channel 满了，延迟关闭连接（需要升级为写锁）
				logx.Errorf("[Hub] Friend %d send buffer full, will close connection", friendInfo.FriendId)
				// 注意：这里在读锁中，不能直接修改 map，通过 Unregister channel 处理
				go func(c *Client) {
					h.Unregister(c)
				}(client)
			}
		}
	}
}

// ==================== 内部实现 ====================

// routeGroupMessage 路由群聊消息的内部实现
// 职责：查询群成员（优先Redis，降级RPC）+ 批量推送给在线成员
func (h *Hub) routeGroupMessage(msg *GroupMessage) {
	var userIds []int64

	// 1. 尝试从 Redis 获取群成员
	redisKey := fmt.Sprintf("im:group:members:%s", msg.GroupId)
	members, err := h.svcCtx.Redis.Smembers(redisKey)
	if err == nil && len(members) > 0 {
		// 缓存命中
		for _, m := range members {
			if uid, err := strconv.ParseInt(m, 10, 64); err == nil {
				userIds = append(userIds, uid)
			}
		}
	} else {
		// 2. 缓存未命中，调用 Group RPC 获取
		ctx := context.Background()
		resp, err := h.svcCtx.GroupRpc.GetMemberList(ctx, &group.GetMemberListReq{
			GroupId:  msg.GroupId,
			Page:     1,
			PageSize: 10000, // 获取所有成员
		})

		if err != nil {
			logx.Errorf("[Hub] Failed to get member list for group %s: %v", msg.GroupId, err)
			return
		}

		// 填充 userIds 并回写 Redis
		var redisMembers []interface{}
		for _, member := range resp.Members {
			userIds = append(userIds, member.UserId)
			redisMembers = append(redisMembers, member.UserId)
		}

		if len(redisMembers) > 0 {
			go func() {
				if _, err := h.svcCtx.Redis.Sadd(redisKey, redisMembers...); err != nil {
					logx.Errorf("[Hub] Failed to cache group members for %s: %v", msg.GroupId, err)
				}
				h.svcCtx.Redis.Expire(redisKey, 7*24*60*60)
			}()
		}
	}

	// 创建排除用户的 map 便于快速查找
	excludeMap := make(map[int64]bool)
	for _, uid := range msg.ExcludeUsers {
		excludeMap[uid] = true
	}

	// 推送消息给所有在线成员（排除指定用户）
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, userId := range userIds {
		// 跳过被排除的用户
		if excludeMap[userId] {
			continue
		}

		// 检查成员是否在线
		if client, ok := h.clients[userId]; ok {
			select {
			case client.send <- msg.Message:
				logx.Infof("[Hub] Sent group message to user %d in group %s", userId, msg.GroupId)
			default:
				// send channel 满了，延迟关闭连接（需要升级为写锁）
				logx.Errorf("[Hub] User %d send buffer full, will close connection", userId)
				// 注意：这里在读锁中，不能直接修改 map，通过 Unregister channel 处理
				go func(c *Client) {
					h.Unregister(c)
				}(client)
			}
		}
	}
}

// mustMarshalMap JSON序列化 map，忽略错误
func mustMarshalMap(v map[string]interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
