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
	"SkyeIM/app/group/rpc/group"
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

// pushOfflineMessages 推送离线消息（优化版：只推最近20条）
func (h *WsHandler) pushOfflineMessages(client *conn.Client) {
	ctx := context.Background()

	// 1. 获取好友列表
	friendResp, err := h.svcCtx.FriendRpc.GetFriendList(ctx, &friend.GetFriendListReq{
		UserId:   client.UserId,
		Page:     1,
		PageSize: 10000,
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
		if friendInfo.Status == 2 {
			continue
		}

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

	// 3. 按时间排序
	sort.Slice(allUnreadMessages, func(i, j int) bool {
		return allUnreadMessages[i].CreatedAt < allUnreadMessages[j].CreatedAt
	})

	// ========== 优化：只推送最近N条 ==========
	const maxPushCount = 20
	totalCount := len(allUnreadMessages)
	pushList := allUnreadMessages

	if totalCount > maxPushCount {
		pushList = allUnreadMessages[totalCount-maxPushCount:]
		logx.Infof("[WsHandler] User %d has %d offline messages, will push latest %d", client.UserId, totalCount, maxPushCount)
	}

	// 4. 发送离线消息摘要
	summaryMsg := &conn.Message{
		Type: "offline_summary",
		Data: mustMarshal(map[string]interface{}{
			"totalCount":  totalCount,
			"pushCount":   len(pushList),
			"hasMore":     totalCount > maxPushCount,
			"remainCount": totalCount - len(pushList),
			"messageType": "private",
		}),
	}

	select {
	case client.SendChannel() <- &conn.BroadcastMessage{Type: "offline_summary", Data: summaryMsg.Data}:
		logx.Infof("[WsHandler] Sent offline summary to user %d: total=%d, push=%d", client.UserId, totalCount, len(pushList))
	default:
		logx.Errorf("[WsHandler] Failed to send offline summary to user %d (buffer full)", client.UserId)
	}

	// 5. 推送消息
	for _, msg := range pushList {
		chatMsg := &conn.ChatMessage{
			MsgId:       msg.MsgId,
			FromUserId:  msg.FromUserId,
			ToUserId:    msg.ToUserId,
			Content:     msg.Content,
			ContentType: msg.ContentType,
			CreatedAt:   msg.CreatedAt,
		}

		wsMsg := &conn.Message{
			Type: "chat",
			Data: mustMarshal(chatMsg),
		}

		select {
		case client.SendChannel() <- &conn.BroadcastMessage{Type: "chat", Data: wsMsg.Data}:
		default:
			logx.Errorf("[WsHandler] Failed to push offline message %s to user %d (buffer full)", msg.MsgId, client.UserId)
		}

		time.Sleep(5 * time.Millisecond)
	}

	logx.Infof("[WsHandler] Successfully pushed %d/%d offline private messages to user %d", len(pushList), totalCount, client.UserId)

	// 6. 推送群聊离线消息
	h.pushOfflineGroupMessages(client)
}

// pushOfflineGroupMessages 推送群聊离线消息
func (h *WsHandler) pushOfflineGroupMessages(client *conn.Client) {
	ctx := context.Background()

	// 1. 获取用户加入的群组列表（包含 ReadSeq）
	groupResp, err := h.svcCtx.GroupRpc.GetJoinedGroups(ctx, &group.GetJoinedGroupsReq{
		UserId: client.UserId,
	})

	if err != nil {
		logx.Errorf("[WsHandler] Failed to get joined groups for user %d: %v", client.UserId, err)
		return
	}

	if len(groupResp.List) == 0 {
		logx.Infof("[WsHandler] User %d has no groups, skip offline group messages", client.UserId)
		return
	}

	// 2. 收集所有群组的未读消息
	var allGroupMessages []*message.MessageInfo

	for _, memberInfo := range groupResp.List {
		// 跳过被禁言的用户？不需要，禁言也能看历史消息
		// 获取大于 ReadSeq 的消息
		msgResp, err := h.svcCtx.MessageRpc.GetGroupMessagesBySeq(ctx, &message.GetGroupMessagesBySeqReq{
			UserId:  client.UserId,
			GroupId: memberInfo.GroupId,
			Seq:     memberInfo.ReadSeq,
		})

		if err != nil {
			logx.Errorf("[WsHandler] Failed to get group messages for user %d in group %s: %v",
				client.UserId, memberInfo.GroupId, err)
			continue
		}

		if len(msgResp.List) > 0 {
			// 过滤掉自己发送的消息（可选，如果自己多端登录可能需要同步给自己）
			// 这里假设同步逻辑: 自己发的消息，其他端虽然已读Seq没更，但通常不需要作为"离线消息"强推，除非为了多端同步。
			// 简单起见，推送所有 > ReadSeq 的消息。
			for _, msg := range msgResp.List {
				if msg.FromUserId != client.UserId {
					allGroupMessages = append(allGroupMessages, msg)
				}
			}
		}
	}

	if len(allGroupMessages) == 0 {
		logx.Infof("[WsHandler] No offline group messages for user %d", client.UserId)
		return
	}

	// 3. 按时间排序（从旧到新），如果时间相同按Seq排序
	sort.Slice(allGroupMessages, func(i, j int) bool {
		if allGroupMessages[i].CreatedAt == allGroupMessages[j].CreatedAt {
			return allGroupMessages[i].Seq < allGroupMessages[j].Seq
		}
		return allGroupMessages[i].CreatedAt < allGroupMessages[j].CreatedAt
	})

	// ========== 优化：只推送最近N条 ==========
	const maxPushCount = 20
	totalCount := len(allGroupMessages)
	pushList := allGroupMessages

	if totalCount > maxPushCount {
		// 只取最后20条（最新的20条）
		pushList = allGroupMessages[totalCount-maxPushCount:]
		logx.Infof("[WsHandler] User %d has %d offline group messages, will push latest %d", client.UserId, totalCount, maxPushCount)
	}

	// 4. 发送群聊离线消息摘要
	summaryMsg := &conn.Message{
		Type: "offline_summary",
		Data: mustMarshal(map[string]interface{}{
			"totalCount":  totalCount,
			"pushCount":   len(pushList),
			"hasMore":     totalCount > maxPushCount,
			"remainCount": totalCount - len(pushList),
			"messageType": "group",
		}),
	}

	select {
	case client.SendChannel() <- &conn.BroadcastMessage{Type: "offline_summary", Data: summaryMsg.Data}:
		logx.Infof("[WsHandler] Sent group offline summary to user %d: total=%d, push=%d", client.UserId, totalCount, len(pushList))
	default:
		logx.Errorf("[WsHandler] Failed to send group offline summary to user %d (buffer full)", client.UserId)
	}

	// 5. 推送群聊消息给客户端
	for _, msg := range pushList {
		// 解析@用户列表
		var atUserIds []int64
		if msg.AtUserIds != nil {
			atUserIds = msg.AtUserIds
		}

		// 检查是否@了当前用户
		isAtMe := false
		for _, id := range atUserIds {
			if id == client.UserId || id == -1 { // -1表示@全体
				isAtMe = true
				break
			}
		}

		groupChatMsg := &conn.GroupChatMessage{
			MsgId:       msg.MsgId,
			FromUserId:  msg.FromUserId,
			GroupId:     msg.GroupId,
			Content:     msg.Content,
			ContentType: msg.ContentType,
			CreatedAt:   msg.CreatedAt,
			Seq:         msg.Seq,
			AtUserIds:   atUserIds,
			IsAtMe:      isAtMe,
		}

		// 构造消息
		wsMsg := &conn.Message{
			Type: "group_chat",
			Data: mustMarshal(groupChatMsg),
		}

		// 发送消息
		select {
		case client.SendChannel() <- &conn.BroadcastMessage{Type: "group_chat", Data: wsMsg.Data}:
			// 发送成功
		default:
			logx.Errorf("[WsHandler] Failed to push offline group message %s to user %d (buffer full)", msg.MsgId, client.UserId)
		}

		// 避免消息推送过快，稍微延迟
		time.Sleep(5 * time.Millisecond)
	}

	logx.Infof("[WsHandler] Successfully pushed %d/%d offline group messages to user %d", len(pushList), totalCount, client.UserId)

	// 6. 不自动更新ReadSeq
	// ReadSeq应该由客户端显式上报，而不是推送后自动标记为已读
	// 推送离线消息不代表用户已读，用户可能还没有真正看到这些消息
	// 客户端应该在用户真正阅读消息后，调用群消息已读上报API
	logx.Infof("[WsHandler] Pushed %d offline group messages to user %d, waiting for client read confirmation", len(pushList), client.UserId)
}

// mustMarshal JSON序列化，忽略错误
func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
