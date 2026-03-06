package conn

// client.go - WebSocket 客户端连接管理
//
// 职责：
// 1. 连接维护：管理单个 WebSocket 连接的生命周期
// 2. 消息读取：ReadPump 从 WebSocket 连接读取消息
// 3. 消息写入：WritePump 向 WebSocket 连接写入消息
// 4. 心跳管理：定期发送 WebSocket Ping 控制帧，处理 Pong 控制帧
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

	// 默认读取超时：服务端在该时间内收不到客户端 Pong，则判定连接失活
	defaultPongWait = 60 * time.Second

	// 默认 Ping 周期：必须小于 Pong 超时，给网络抖动留出缓冲
	defaultPingPeriod = (defaultPongWait * 9) / 10

	// 默认最大消息大小
	defaultMaxMessageSize = 65536
)

// Client 代表一个 WebSocket 客户端连接
type Client struct {
	Hub    *Hub
	UserId int64
	conn   *websocket.Conn
	send   chan interface{}
	svcCtx *svc.ServiceContext

	// 心跳配置（协议层 WebSocket 控制帧）
	// Server -> Client: Ping
	// Client -> Server: Pong（由客户端 WebSocket 库自动回复）
	pongWait   time.Duration
	pingPeriod time.Duration

	// 读取限制
	maxMessageSize int64

	done      chan struct{}
	closeOnce sync.Once
}

// NewClient 创建新的客户端
func NewClient(hub *Hub, conn *websocket.Conn, userId int64, svcCtx *svc.ServiceContext) *Client {
	pongWait := defaultPongWait
	pingPeriod := defaultPingPeriod
	maxMessageSize := int64(defaultMaxMessageSize)

	// 使用配置覆盖默认值，保持配置与运行时行为一致
	if svcCtx != nil {
		wsCfg := svcCtx.Config.WebSocket
		if wsCfg.PongTimeout > 0 {
			pongWait = time.Duration(wsCfg.PongTimeout) * time.Second
		}
		if wsCfg.PingInterval > 0 {
			pingPeriod = time.Duration(wsCfg.PingInterval) * time.Second
		}
		if wsCfg.MaxMessageSize > 0 {
			maxMessageSize = wsCfg.MaxMessageSize
		}
	}

	// 兜底保护：Ping 周期必须小于 Pong 超时，避免“刚发 Ping 就超时断开”
	if pingPeriod >= pongWait {
		pingPeriod = (pongWait * 9) / 10
	}

	return &Client{
		Hub:            hub,
		UserId:         userId,
		conn:           conn,
		send:           make(chan interface{}, 256),
		svcCtx:         svcCtx,
		pongWait:       pongWait,
		pingPeriod:     pingPeriod,
		maxMessageSize: maxMessageSize,
		done:           make(chan struct{}),
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

	c.conn.SetReadLimit(c.maxMessageSize)
	// 读取超时窗口：超过该时间未收到任何数据/控制帧（含 Pong）将触发断开
	c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
	// 收到客户端 Pong 控制帧时续期读超时
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(c.pongWait))
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
	// 服务端定时发送 WebSocket Ping 控制帧，驱动链路保活
	ticker := time.NewTicker(c.pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case <-c.done:
			return

		case message, ok := <-c.send:
			if !ok {
				// send channel 已被关闭，发送 WS 关闭帧后退出
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
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
