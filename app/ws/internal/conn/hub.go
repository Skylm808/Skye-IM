package conn

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"SkyeIM/app/friend/rpc/friend"
	"SkyeIM/app/ws/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// mustMarshalMap JSON序列化 map，忽略错误
func mustMarshalMap(v map[string]interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

// Hub 维护活跃的客户端连接集合
type Hub struct {
	// 在线用户映射: userId -> Client
	clients map[int64]*Client

	// 注册请求通道
	register chan *Client

	// 注销请求通道
	unregister chan *Client

	// 广播消息通道
	broadcast chan *BroadcastMessage

	// 私聊消息通道
	private chan *PrivateMessage

	// 服务上下文（用于调用 RPC）
	svcCtx *svc.ServiceContext

	// 互斥锁
	mu sync.RWMutex
}

// BroadcastMessage 广播消息
type BroadcastMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// PrivateMessage 私聊消息
type PrivateMessage struct {
	ToUserId int64
	Message  *Message
}

// NewHub 创建新的Hub
func NewHub(svcCtx *svc.ServiceContext) *Hub {
	return &Hub{
		clients:    make(map[int64]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMessage),
		private:    make(chan *PrivateMessage, 256),
		svcCtx:     svcCtx,
	}
}

// Run 启动Hub的消息循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// 如果已有连接，先关闭旧连接
			if oldClient, ok := h.clients[client.UserId]; ok {
				close(oldClient.send)
				oldClient.conn.Close()
			}
			h.clients[client.UserId] = client
			h.mu.Unlock()
			logx.Infof("[Hub] User %d connected, total online: %d", client.UserId, len(h.clients))

			// 通知该用户的好友上线
			h.notifyOnlineStatus(client.UserId, true)

		case client := <-h.unregister:
			h.mu.Lock()
			if c, ok := h.clients[client.UserId]; ok && c == client {
				delete(h.clients, client.UserId)
				close(client.send)
			}
			h.mu.Unlock()
			logx.Infof("[Hub] User %d disconnected, total online: %d", client.UserId, len(h.clients))

			// 通知该用户的好友下线
			h.notifyOnlineStatus(client.UserId, false)

		case msg := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.send <- msg:
				default:
					// 发送失败，关闭连接
					close(client.send)
					delete(h.clients, client.UserId)
				}
			}
			h.mu.RUnlock()

		case msg := <-h.private:
			h.mu.RLock()
			if client, ok := h.clients[msg.ToUserId]; ok {
				data, _ := json.Marshal(msg.Message)
				select {
				case client.send <- &BroadcastMessage{Type: msg.Message.Type, Data: json.RawMessage(data)}:
				default:
					logx.Errorf("[Hub] Failed to send message to user %d", msg.ToUserId)
				}
			}
			h.mu.RUnlock()
		}
	}
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
				logx.Errorf("[Hub] Failed to notify friend %d (send buffer full)", friendInfo.FriendId)
			}
		}
	}
}

// SendToUser 发送消息给指定用户
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
		logx.Errorf("[Hub] Failed to send message to user %d: send buffer full", userId)
		return false
	}
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

// Register 注册客户端
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister 注销客户端
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}
