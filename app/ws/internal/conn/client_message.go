package conn

// client_message.go - Client 消息处理业务逻辑
//
// 职责：
// 1. 消息解析：解析客户端发来的各类消息（私聊、群聊、ACK、已读）
// 2. 业务验证：验证消息合法性、权限检查
// 3. 数据存储：调用 RPC 将消息持久化到数据库
// 4. ACK 确认：向发送者返回消息确认（sent/failed）
// 5. 路由请求：调用 Hub 的路由方法分发消息
//
// 设计说明：
// - 本文件专注于业务逻辑，不关心"如何路由"
// - 路由的具体实现在 hub.go 中
// - 私聊调用 Hub.SendToUser()（同步）
// - 群聊调用 Hub.SendToGroup()（异步）

import (
	"context"
	"encoding/json"
	"time"

	"SkyeIM/app/group/rpc/group"
	"SkyeIM/app/message/rpc/message"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

func (c *Client) sendAck(msgId string, status string, reason string, timestamp int64) {
	ackMsg := &Message{
		Type: "ack",
		Data: mustMarshal(&AckMessage{
			MsgId:     msgId,
			Status:    status,
			Reason:    reason,
			Timestamp: timestamp,
		}),
	}

	select {
	case c.send <- ackMsg:
	default:
		logx.Errorf("[Client] Failed to send ACK (%s) to user %d: send buffer full", status, c.UserId)
	}
}

func (c *Client) sendError(msgId string, message string) {
	errMsg := &Message{
		Type: "error",
		Data: mustMarshal(map[string]interface{}{
			"msgId":   msgId,
			"message": message,
		}),
	}

	select {
	case c.send <- errMsg:
	default:
		logx.Errorf("[Client] Failed to send error to user %d: send buffer full", c.UserId)
	}
}

// handleChatMessage 处理私聊消息
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
		c.sendAck(chatMsg.MsgId, "failed", "rpc_error", time.Now().Unix())
		c.sendError(chatMsg.MsgId, "发送失败")
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

// handleGroupChatMessage 处理群聊消息
func (c *Client) handleGroupChatMessage(data json.RawMessage) {
	var groupMsg GroupChatMessage
	if err := json.Unmarshal(data, &groupMsg); err != nil {
		logx.Errorf("[Client] User %d parse group chat message error: %v", c.UserId, err)
		return
	}

	// 设置发送者ID
	groupMsg.FromUserId = c.UserId

	// 生成消息ID（如果客户端未提供）
	if groupMsg.MsgId == "" {
		groupMsg.MsgId = uuid.New().String()
	}

	// 默认消息类型为文字
	if groupMsg.ContentType == 0 {
		groupMsg.ContentType = 1
	}

	ctx := context.Background()

	// 调用 Group RPC 验证用户是否在群组中
	checkResp, err := c.svcCtx.GroupRpc.CheckMembership(ctx, &group.CheckMembershipReq{
		GroupId: groupMsg.GroupId,
		UserId:  c.UserId,
	})

	if err != nil {
		logx.Errorf("[Client] CheckMembership failed for user %d in group %s: %v", c.UserId, groupMsg.GroupId, err)
		c.sendAck(groupMsg.MsgId, "failed", "check_failed", time.Now().Unix())
		c.sendError(groupMsg.MsgId, "群成员校验失败")
		return
	}

	if !checkResp.IsMember {
		logx.Errorf("[Client] User %d is not a member of group %s", c.UserId, groupMsg.GroupId)
		c.sendAck(groupMsg.MsgId, "failed", "not_member", time.Now().Unix())
		c.sendError(groupMsg.MsgId, "您不是该群组成员")
		return
	}

	// 检查是否被禁言
	if checkResp.Member.Mute == 1 {
		logx.Errorf("[Client] User %d is muted in group %s", c.UserId, groupMsg.GroupId)
		c.sendAck(groupMsg.MsgId, "failed", "muted", time.Now().Unix())
		c.sendError(groupMsg.MsgId, "您已被禁言")
		return
	}

	// 存储群聊消息到数据库
	resp, err := c.svcCtx.MessageRpc.SendGroupMessage(ctx, &message.SendGroupMessageReq{
		MsgId:       groupMsg.MsgId,
		FromUserId:  groupMsg.FromUserId,
		GroupId:     groupMsg.GroupId,
		Content:     groupMsg.Content,
		ContentType: groupMsg.ContentType,
		AtUserIds:   groupMsg.AtUserIds,
	})

	if err != nil {
		logx.Errorf("[Client] User %d send group message failed: %v", c.UserId, err)
		c.sendAck(groupMsg.MsgId, "failed", "rpc_error", time.Now().Unix())
		c.sendError(groupMsg.MsgId, "发送失败")
		return
	}

	// 更新时间戳和Seq
	groupMsg.CreatedAt = resp.CreatedAt
	groupMsg.Seq = resp.Seq

	// 发送 ACK 给发送者
	ackMsg := &Message{
		Type: "ack",
		Data: mustMarshal(&AckMessage{
			MsgId:     groupMsg.MsgId,
			Status:    "sent",
			Timestamp: resp.CreatedAt,
		}),
	}

	select {
	case c.send <- ackMsg:
		logx.Infof("[Client] Sent ACK (sent) to user %d for group message %s", c.UserId, groupMsg.MsgId)
	default:
		logx.Errorf("[Client] Failed to send ACK to user %d: send buffer full", c.UserId)
	}

	// 构造发送给群成员的消息
	groupReceiverMsg := &Message{
		Type: "group_chat",
		Data: mustMarshal(&groupMsg),
	}

	// 推送给群组所有在线成员（排除发送者自己）
	c.Hub.SendToGroup(groupMsg.GroupId, groupReceiverMsg, []int64{c.UserId})

	logx.Infof("[Client] Group message %s sent to group %s by user %d", groupMsg.MsgId, groupMsg.GroupId, c.UserId)
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
