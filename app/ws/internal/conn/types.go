package conn

import "encoding/json"

// Message WebSocket消息格式
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ChatMessage 私聊消息数据
type ChatMessage struct {
	MsgId       string `json:"msgId,omitempty"`
	FromUserId  int64  `json:"fromUserId"`
	ToUserId    int64  `json:"toUserId"`
	Content     string `json:"content"`
	ContentType int32  `json:"contentType"`
	CreatedAt   int64  `json:"createdAt,omitempty"`
}

// GroupChatMessage 群聊消息数据
type GroupChatMessage struct {
	MsgId       string  `json:"msgId,omitempty"`
	FromUserId  int64   `json:"fromUserId"`
	GroupId     string  `json:"groupId"`
	Content     string  `json:"content"`
	ContentType int32   `json:"contentType"`
	CreatedAt   int64   `json:"createdAt,omitempty"`
	Seq         uint64  `json:"seq,omitempty"`
	AtUserIds   []int64 `json:"atUserIds,omitempty"` // 被@的用户ID列表，-1表示@全体
	IsAtMe      bool    `json:"isAtMe,omitempty"`    // 是否@了当前用户
}

// AckMessage 确认消息
type AckMessage struct {
	MsgId     string `json:"msgId"`
	Status    string `json:"status"` // sent, delivered, read, failed
	Reason    string `json:"reason,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// GroupMessage 群组消息通道数据
type GroupMessage struct {
	GroupId      string
	FromUserId   int64
	Message      *Message
	ExcludeUsers []int64 // 排除的用户列表（如发送者自己）
}

// ==================== 群组事件数据结构 ====================

// GroupMemberJoinEvent 成员加入事件
type GroupMemberJoinEvent struct {
	GroupId   string `json:"groupId"`
	UserId    int64  `json:"userId"`
	InviterId int64  `json:"inviterId"`
	Timestamp int64  `json:"timestamp"`
}

// GroupMemberLeaveEvent 成员退出事件
type GroupMemberLeaveEvent struct {
	GroupId   string `json:"groupId"`
	UserId    int64  `json:"userId"`
	Timestamp int64  `json:"timestamp"`
}

// GroupMemberKickEvent 成员被踢事件
type GroupMemberKickEvent struct {
	GroupId    string `json:"groupId"`
	UserId     int64  `json:"userId"`
	OperatorId int64  `json:"operatorId"`
	Timestamp  int64  `json:"timestamp"`
}

// GroupInfoUpdateEvent 群组信息更新事件
type GroupInfoUpdateEvent struct {
	GroupId    string `json:"groupId"`
	OperatorId int64  `json:"operatorId"`
	UpdateType string `json:"updateType"` // name/avatar/description
	NewValue   string `json:"newValue"`
	Timestamp  int64  `json:"timestamp"`
}

// GroupDismissEvent 群组解散事件
type GroupDismissEvent struct {
	GroupId    string `json:"groupId"`
	OperatorId int64  `json:"operatorId"`
	Timestamp  int64  `json:"timestamp"`
}
