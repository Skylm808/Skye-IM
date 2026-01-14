package conn

// client.go - WebSocket 客户端连接管理
//
// 职责：
// 1. 连接维护：管理单个 WebSocket 连接的生命周期
// 2. 消息读取：ReadPump 从 WebSocket 连接读取消息
// 3. 消息写入：WritePump 向 WebSocket 连接写入消息
// 4. 心跳管理：定期发送 Ping，处理 Pong
// 5. 消息分发：将收到的消息路由到对应的处理函数
//
// 设计说明：
// - 一个 Client 对应一个 WebSocket 连接
// - ReadPump 和 WritePump 各自在独立的 goroutine 中运行
// - send channel 用于异步发送消息给客户端

import (
	"encoding/json"
	"sync"
	"time"

	"SkyeIM/app/ws/internal/svc"

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

	done      chan struct{}
	closeOnce sync.Once
}

// NewClient 创建新的客户端
func NewClient(hub *Hub, conn *websocket.Conn, userId int64, svcCtx *svc.ServiceContext) *Client {
	return &Client{
		Hub:    hub,
		UserId: userId,
		conn:   conn,
		send:   make(chan interface{}, 256),
		svcCtx: svcCtx,
		done:   make(chan struct{}),
	}
}

// Close 关闭连接并停止读写协程（幂等）
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.done)
		if c.conn != nil {
			_ = c.conn.Close()
		}
	})
}

// SendChannel 返回发送通道（用于外部推送消息）
func (c *Client) SendChannel() chan interface{} {
	return c.send
}

// ReadPump 读取消息的协程
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Close()
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
		c.Close()
	}()

	for {
		select {
		case <-c.done:
			return

		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			_ = ok

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

// handleMessage 处理收到的消息（路由到具体处理函数）
func (c *Client) handleMessage(msg *Message) {
	switch msg.Type {
	case "ping":
		// 心跳响应
		c.send <- &Message{Type: "pong", Data: nil}

	case "chat":
		// 处理私聊消息
		c.handleChatMessage(msg.Data)

	case "group_chat":
		// 处理群聊消息
		c.handleGroupChatMessage(msg.Data)

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
