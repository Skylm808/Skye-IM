package conn

import (
	"context"
	"encoding/json"
	"time"

	"SkyeIM/app/message/rpc/message"
	"SkyeIM/app/ws/internal/svc"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// 写入等待时间
	writeWait = 10 * time.Second

	// 读取 pong 超时时间
	pongWait = 60 * time.Second

	// 发送 ping 的周期
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 65536
)

// Client 代表一个 WebSocket 客户端连接
type Client struct {
	Hub    *Hub
	UserId int64
	conn   *websocket.Conn
	send   chan interface{}
	svcCtx *svc.ServiceContext
}

// Message WebSocket消息格式
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ChatMessage 聊天消息数据
type ChatMessage struct {
	MsgId       string `json:"msgId,omitempty"`
	FromUserId  int64  `json:"fromUserId"`
	ToUserId    int64  `json:"toUserId"`
	Content     string `json:"content"`
	ContentType int32  `json:"contentType"`
	CreatedAt   int64  `json:"createdAt,omitempty"`
}

// AckMessage 确认消息
type AckMessage struct {
	MsgId     string `json:"msgId"`
	Status    string `json:"status"` // sent, delivered, read
	Timestamp int64  `json:"timestamp"`
}

// NewClient 创建新的客户端
func NewClient(hub *Hub, conn *websocket.Conn, userId int64, svcCtx *svc.ServiceContext) *Client {
	return &Client{
		Hub:    hub,
		UserId: userId,
		conn:   conn,
		send:   make(chan interface{}, 256),
		svcCtx: svcCtx,
	}
}

// SendChannel 返回发送通道（用于外部推送消息）
func (c *Client) SendChannel() chan interface{} {
	return c.send
}

// ReadPump 读取消息的协程
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msgBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Errorf("[Client] User %d read error: %v", c.UserId, err)
			}
			break
		}

		// 解析消息
		var msg Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			logx.Errorf("[Client] User %d parse message error: %v", c.UserId, err)
			continue
		}

		// 处理消息
		c.handleMessage(&msg)
	}
}

// WritePump 写入消息的协程
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub 关闭了通道
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				logx.Errorf("[Client] User %d write error: %v", c.UserId, err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理收到的消息
func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case "ping":
		// 心跳响应
		c.send <- &Message{Type: "pong", Data: nil}

	case "chat":
		// 处理聊天消息
		c.handleChatMessage(msg.Data)

	case "ack":
		// 处理消息确认
		c.handleAckMessage(msg.Data)

	case "read":
		// 处理已读回执
		c.handleReadMessage(msg.Data)

	default:
		logx.Infof("[Client] User %d unknown message type: %s", c.UserId, msg.Type)
	}
}

// handleChatMessage 处理聊天消息
func (c *Client) handleChatMessage(data json.RawMessage) {
	var chatMsg ChatMessage
	if err := json.Unmarshal(data, &chatMsg); err != nil {
		logx.Errorf("[Client] User %d parse chat message error: %v", c.UserId, err)
		return
	}

	// 设置发送者ID
	chatMsg.FromUserId = c.UserId

	// 生成消息ID（如果客户端未提供）
	if chatMsg.MsgId == "" {
		chatMsg.MsgId = uuid.New().String()
	}

	// 默认消息类型为文字
	if chatMsg.ContentType == 0 {
		chatMsg.ContentType = 1
	}

	// 存储消息到数据库
	ctx := context.Background()
	resp, err := c.svcCtx.MessageRpc.SendMessage(ctx, &message.SendMessageReq{
		MsgId:       chatMsg.MsgId,
		FromUserId:  chatMsg.FromUserId,
		ToUserId:    chatMsg.ToUserId,
		Content:     chatMsg.Content,
		ContentType: chatMsg.ContentType,
	})

	if err != nil {
		logx.Errorf("[Client] User %d send message failed: %v", c.UserId, err)
		// 发送错误回复给发送者
		c.send <- &Message{
			Type: "error",
			Data: mustMarshal(map[string]interface{}{
				"msgId":   chatMsg.MsgId,
				"message": "发送失败",
			}),
		}
		return
	}

	// 更新时间戳
	chatMsg.CreatedAt = resp.CreatedAt

	// 发送 ACK 给发送者（非阻塞）
	ackMsg := &Message{
		Type: "ack",
		Data: mustMarshal(&AckMessage{
			MsgId:     chatMsg.MsgId,
			Status:    "sent",
			Timestamp: resp.CreatedAt,
		}),
	}

	select {
	case c.send <- ackMsg:
		logx.Infof("[Client] Sent ACK (sent) to user %d for message %s", c.UserId, chatMsg.MsgId)
	default:
		logx.Errorf("[Client] Failed to send ACK to user %d: send buffer full", c.UserId)
	}

	// 构造发送给接收者的消息
	receiverMsg := &Message{
		Type: "chat",
		Data: mustMarshal(&chatMsg),
	}

	// 尝试发送给接收者
	if c.Hub.SendToUser(chatMsg.ToUserId, receiverMsg) {
		// 接收者在线，发送已送达确认给发送者
		deliveredAck := &Message{
			Type: "ack",
			Data: mustMarshal(&AckMessage{
				MsgId:     chatMsg.MsgId,
				Status:    "delivered",
				Timestamp: time.Now().Unix(),
			}),
		}

		select {
		case c.send <- deliveredAck:
			logx.Infof("[Client] Sent ACK (delivered) to user %d for message %s", c.UserId, chatMsg.MsgId)
		default:
			logx.Errorf("[Client] Failed to send delivered ACK to user %d: send buffer full", c.UserId)
		}
	} else {
		logx.Infof("[Client] User %d is offline, message %s stored for later delivery", chatMsg.ToUserId, chatMsg.MsgId)
	}
	// 如果接收者不在线，消息已存储在数据库，下次上线时会推送
}

// handleAckMessage 处理消息确认
func (c *Client) handleAckMessage(data json.RawMessage) {
	var ack AckMessage
	if err := json.Unmarshal(data, &ack); err != nil {
		logx.Errorf("[Client] User %d parse ack message error: %v", c.UserId, err)
		return
	}
	logx.Infof("[Client] User %d ack message: %s, status: %s", c.UserId, ack.MsgId, ack.Status)
}

// handleReadMessage 处理已读回执
func (c *Client) handleReadMessage(data json.RawMessage) {
	var readMsg struct {
		PeerId int64    `json:"peerId"`
		MsgIds []string `json:"msgIds,omitempty"`
	}
	if err := json.Unmarshal(data, &readMsg); err != nil {
		logx.Errorf("[Client] User %d parse read message error: %v", c.UserId, err)
		return
	}

	// 标记消息为已读
	ctx := context.Background()
	_, err := c.svcCtx.MessageRpc.MarkAsRead(ctx, &message.MarkAsReadReq{
		UserId: c.UserId,
		PeerId: readMsg.PeerId,
		MsgIds: readMsg.MsgIds,
	})

	if err != nil {
		logx.Errorf("[Client] User %d mark as read failed: %v", c.UserId, err)
		return
	}

	// 通知对方消息已读
	c.Hub.SendToUser(readMsg.PeerId, &Message{
		Type: "read",
		Data: mustMarshal(map[string]interface{}{
			"userId":    c.UserId,
			"timestamp": time.Now().Unix(),
		}),
	})
}

// mustMarshal JSON序列化，忽略错误
func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
