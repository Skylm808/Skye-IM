package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"time"

	"SkyeIM/app/friend/rpc/friend"
	"SkyeIM/app/message/rpc/message"
	"SkyeIM/app/ws/internal/config"
	"SkyeIM/app/ws/internal/conn"
	"SkyeIM/app/ws/internal/svc"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许跨域
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WsHandler WebSocket连接处理器
type WsHandler struct {
	svcCtx *svc.ServiceContext
	hub    *conn.Hub
	config config.Config
}

// NewWsHandler 创建WebSocket处理器
func NewWsHandler(svcCtx *svc.ServiceContext, hub *conn.Hub) *WsHandler {
	return &WsHandler{
		svcCtx: svcCtx,
		hub:    hub,
		config: svcCtx.Config,
	}
}

// ServeHTTP 处理WebSocket连接请求
func (h *WsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 从 URL 参数或 Header 获取 token
	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}
	}

	// 验证 token
	userId, err := h.parseToken(token)
	if err != nil {
		logx.Errorf("[WsHandler] Token validation failed: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 升级为 WebSocket 连接
	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Errorf("[WsHandler] WebSocket upgrade failed: %v", err)
		return
	}

	// 创建客户端
	client := conn.NewClient(h.hub, wsConn, userId, h.svcCtx)

	// 注册到 Hub
	h.hub.Register(client)

	// 发送连接成功消息
	wsConn.WriteJSON(map[string]interface{}{
		"type": "connected",
		"data": map[string]interface{}{
			"userId":      userId,
			"onlineCount": h.hub.OnlineCount(),
		},
	})

	// 推送离线消息
	go h.pushOfflineMessages(client)

	// 启动读写协程
	go client.WritePump()
	go client.ReadPump()
}

// parseToken 解析并验证JWT Token
func (h *WsHandler) parseToken(tokenString string) (int64, error) {
	if tokenString == "" {
		return 0, errors.New("token is empty")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.config.Auth.AccessSecret), nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}

	// 获取用户ID (可能是 "userId" 或 "sub")
	var userId int64
	if uid, ok := claims["userId"]; ok {
		switch v := uid.(type) {
		case float64:
			userId = int64(v)
		case string:
			userId, _ = strconv.ParseInt(v, 10, 64)
		}
	}

	if userId == 0 {
		return 0, errors.New("userId not found in token")
	}

	return userId, nil
}

// pushOfflineMessages 推送离线消息
func (h *WsHandler) pushOfflineMessages(client *conn.Client) {
	ctx := context.Background()

	// 1. 获取好友列表
	friendResp, err := h.svcCtx.FriendRpc.GetFriendList(ctx, &friend.GetFriendListReq{
		UserId:   client.UserId,
		Page:     1,
		PageSize: 10000, // 获取所有好友
	})

	if err != nil {
		logx.Errorf("[WsHandler] Failed to get friend list for user %d: %v", client.UserId, err)
		return
	}

	if len(friendResp.List) == 0 {
		logx.Infof("[WsHandler] User %d has no friends, skip offline messages", client.UserId)
		return
	}

	// 2. 收集所有未读消息
	var allUnreadMessages []*message.MessageInfo

	for _, friendInfo := range friendResp.List {
		// 跳过被拉黑的好友
		if friendInfo.Status == 2 {
			continue
		}

		// 获取与该好友的未读消息
		unreadResp, err := h.svcCtx.MessageRpc.GetUnreadMessages(ctx, &message.GetUnreadMessagesReq{
			UserId: client.UserId,
			PeerId: friendInfo.FriendId,
		})

		if err != nil {
			logx.Errorf("[WsHandler] Failed to get unread messages from user %d: %v", friendInfo.FriendId, err)
			continue
		}

		if len(unreadResp.List) > 0 {
			allUnreadMessages = append(allUnreadMessages, unreadResp.List...)
			logx.Infof("[WsHandler] Found %d unread messages from user %d", len(unreadResp.List), friendInfo.FriendId)
		}
	}

	if len(allUnreadMessages) == 0 {
		logx.Infof("[WsHandler] No offline messages for user %d", client.UserId)
		return
	}

	// 3. 按时间排序（从旧到新）
	sort.Slice(allUnreadMessages, func(i, j int) bool {
		return allUnreadMessages[i].CreatedAt < allUnreadMessages[j].CreatedAt
	})

	logx.Infof("[WsHandler] Pushing %d offline messages to user %d", len(allUnreadMessages), client.UserId)

	// 4. 推送消息给客户端
	for _, msg := range allUnreadMessages {
		chatMsg := &conn.ChatMessage{
			MsgId:       msg.MsgId,
			FromUserId:  msg.FromUserId,
			ToUserId:    msg.ToUserId,
			Content:     msg.Content,
			ContentType: msg.ContentType,
			CreatedAt:   msg.CreatedAt,
		}

		// 构造消息
		wsMsg := &conn.Message{
			Type: "chat",
			Data: mustMarshal(chatMsg),
		}

		// 发送消息
		select {
		case client.SendChannel() <- &conn.BroadcastMessage{Type: "chat", Data: wsMsg.Data}:
			// 发送成功
		default:
			logx.Errorf("[WsHandler] Failed to push offline message %s to user %d (buffer full)", msg.MsgId, client.UserId)
		}

		// 避免消息推送过快，稍微延迟
		time.Sleep(10 * time.Millisecond)
	}

	logx.Infof("[WsHandler] Successfully pushed %d offline messages to user %d", len(allUnreadMessages), client.UserId)
}

// mustMarshal JSON序列化，忽略错误
func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
