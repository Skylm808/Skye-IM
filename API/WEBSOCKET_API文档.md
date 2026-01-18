# WebSocket 服务对接文档

## 📋 目录

- [概述](#概述)
- [连接建立](#连接建立)
- [消息格式](#消息格式)
- [心跳机制](#心跳机制)
- [离线消息推送](#离线消息推送)
- [前端事件处理指南 (新增)](#前端事件处理指南)
- [错误处理](#错误处理)
- [常见问题](#常见问题)

---

## 概述

### 服务地址

```
WebSocket: ws://localhost:10300/ws
健康检查: http://localhost:10300/health
```

### 核心功能

| 功能 | 说明 |
|------|------|
| 实时消息 | 收发私聊和群聊消息 |
| 在线状态 | 维护用户在线状态 |
| 离线推送 | 上线时推送离线消息（前20条） |
| 心跳保活 | 30秒心跳，保持连接 |
| 事件通知 | 好友请求、群组邀请等 |

---

## 连接建立

### 1. 连接URL

```
ws://localhost:10300/ws?token=<JWT_ACCESS_TOKEN>
```

**参数说明**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|-----|------|
| token | string | 是 | JWT Access Token（从登录接口获取） |

**请求示例**:
```
ws://localhost:10300/ws?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 2. 连接流程

```
1. 客户端发起 WebSocket 连接
   ↓
2. 服务端验证 JWT Token
   ↓ 验证通过
3. 连接成功，服务端分配 Connection ID
   ↓
4. 自动推送离线消息（前20条）
   ↓
5. 开始心跳
```

### 3. 连接成功响应

连接成功后，服务端会立即推送欢迎消息：

```json
{
  "type": "connected",
  "data": {
    "userId": 1001,
    "connectedAt": 1736683200
  }
}
```

### 4. 连接失败

**情况一：Token 无效**
```
WebSocket连接立即关闭
关闭码: 1008 (Policy Violation)
原因: "Invalid token"
```

**情况二：Token 过期**
```
关闭码: 1008
原因: "Token expired"
```

**处理方式**:
1. 刷新 Token（调用 `/api/v1/auth/refresh`）
2. 使用新 Token 重新连接

---

## 消息格式

### 消息结构

所有 WebSocket 消息都使用 JSON 格式：

```json
{
  "type": "消息类型",
  "data": { /* 消息内容 */ }
}
```

### 消息类型

| type | 方向 | 说明 |
|------|------|------|
| `ping` | 客户端→服务端 | 心跳请求 |
| `pong` | 服务端→客户端 | 心跳响应 |
| `chat` | 双向 | 私聊消息 |
| `group_chat` | 双向 | 群聊消息 |
| `connected` | 服务端→客户端 | 连接成功 |
| `friend_request` | 服务端→客户端 | 好友请求通知 |
| `group_invitation` | 服务端→客户端 | 群组邀请通知 |
| `group_event` | 服务端→客户端 | 群组变更通知 (解散/入群/退群等) |
| `read` | 服务端→客户端 | 已读回执 |
| `offline_messages` | 服务端→客户端 | 离线消息摘要通知 |

---

### 1. 发送私聊消息

**客户端发送**:
```json
{
  "type": "chat",
  "data": {
    "toUserId": 1002,
    "content": "你好，在吗？",
    "contentType": 1,
    "msgId": "msg_client_generated_id"
  }
}
```

**字段说明**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|-----|------|
| toUserId | int64 | 是 | 接收者用户ID |
| content | string | 是 | 消息内容 |
| contentType | int32 | 是 | 内容类型：1-文本 2-图片 3-文件 4-语音 |
| msgId | string | 否 | 客户端生成的消息ID（用于去重） |

**服务端响应（发送成功）**:
```json
{
  "type": "chat",
  "data": {
    "id": 12345,
    "msgId": "msg_20260113_12345",
    "fromUserId": 1001,
    "toUserId": 1002,
    "content": "你好，在吗？",
    "contentType": 1,
    "status": 0,
    "createdAt": 1736683200
  }
}
```

**接收者收到的消息**:
```json
{
  "type": "chat",
  "data": {
    "id": 12345,
    "msgId": "msg_20260113_12345",
    "fromUserId": 1001,
    "toUserId": 1002,
    "content": "你好，在吗？",
    "contentType": 1,
    "status": 0,
    "createdAt": 1736683200
  }
}
```

---

### 2. 发送群聊消息

**客户端发送**:
```json
{
  "type": "group_chat",
  "data": {
    "groupId": "g_20260113_001",
    "content": "@张三 明天开会",
    "contentType": 1,
    "atUserIds": [1002],
    "msgId": "msg_client_generated_id"
  }
}
```

**字段说明**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|-----|------|
| groupId | string | 是 | 群组ID |
| content | string | 是 | 消息内容 |
| contentType | int32 | 是 | 内容类型：1-文本 2-图片 3-文件 4-语音 |
| atUserIds | []int64 | 否 | 被@的用户ID列表，-1表示@全体成员 |
| msgId | string | 否 | 客户端生成的消息ID |

**服务端响应**:
```json
{
  "type": "group_chat",
  "data": {
    "id": 12350,
    "msgId": "msg_20260113_12350",
    "fromUserId": 1001,
    "groupId": "g_20260113_001",
    "content": "@张三 明天开会",
    "contentType": 1,
    "status": 0,
    "createdAt": 1736683300,
    "seq": 1250,
    "atUserIds": [1002]
  }
}
```

**群内其他成员收到的消息**:
格式相同，所有在线成员都会收到。

---

### 3. 接收离线消息

**连接成功后自动推送**:
```json
{
  "type": "offline_messages",
  "data": {
    "messages": [
      {
        "id": 12340,
        "msgId": "msg_20260113_12340",
        "fromUserId": 1003,
        "toUserId": 1001,
        "content": "晚上一起吃饭吗？",
        "contentType": 1,
        "status": 0,
        "createdAt": 1736683100
      }
    ],
    "totalCount": 25,
    "hasMore": true,
    "messageType": "private"  // 或 "group"
  }
}
```

**字段说明**:
| 字段 | 类型 | 说明 |
|------|------|------|
| messages | array | 离线消息列表（前20条） |
| totalCount | int64 | 总离线消息数 |
| hasMore | bool | 是否还有更多（true时需调用HTTP接口拉取） |
| messageType | string | 消息类型：private(私聊) / group(群聊) |

**拉取剩余离线消息**:
如果 `hasMore=true`，调用 HTTP 接口：
```
GET /api/v1/message/offline?skip=20&limit=100
```

---

### 4. 接收事件通知

**好友请求通知**:
```json
{
  "type": "friend_request",
  "data": {
    "id": 123,
    "fromUserId": 1002,
    "fromUserName": "李四",
    "message": "我是李四，想加你为好友",
    "createdAt": 1736683200
  }
}
```

**群组邀请通知**:
```json
{
  "type": "group_invitation",
  "data": {
    "invitationId": 456,
    "groupId": "g_20260113_001",
    "groupName": "技术交流群",
    "inviterId": 1003,
    "inviterName": "王五",
    "message": "来我们群聊聊天吧",
    "createdAt": 1736683300
  }
}
```

**已读回执**:
```json
{
  "type": "read",
  "data": {
    "msgIds": ["msg_20260113_12345", "msg_20260113_12346"],
    "readBy": 1002,
    "readAt": 1736683400
  }
}
```

---

## 前端事件处理指南

本节详细说明收到各类 `group_event` 时的推荐处理逻辑。

**消息格式**:
```json
{
  "type": "group_event",
  "eventType": "dismissGroup", // 事件类型
  "eventData": { ... } // 数据
}
```

### 1. 群组解散 (dismissGroup)

**触发时机**: 群主解散群聊。
**数据示例**: `{"groupId": "1001", "operatorId": 888}`

**前端动作**:
1. 弹出提示："群聊 [1001] 已被解散"。
2. **状态更新**：从本地 Store (Redux/Vuex) 的群组列表中**移除**该群。
3. **界面跳转**：如果用户当前正停留在该群的聊天界面，强制跳转回首页或空白页。

### 2. 成员被踢 (kickMember)

**触发时机**: 管理员将成员移出群聊。
**数据示例**: `{"operatorId": 888, "memberId": 999, "groupId": "1001"}`

**前端动作**:
*   **如果不幸是你 (`memberId == currentUser.id`)**:
    1. 弹出提示："你已被移出群聊 [1001]"。
    2. **状态更新**：从本地 Store 的群组列表中**移除**该群。
    3. **界面跳转**：如果正在该群聊天，强制退出。
*   **如果是别人被踢**:
    1. **状态更新**：从该群的"成员列表"中移除该用户。
    2. (可选) 在聊天窗口插入一条系统消息："用户 [999] 被移出群聊"。

### 3. 主动退群 (quitGroup)

**触发时机**: 成员主动退出。
**数据示例**: `{"userId": 999, "groupId": "1001"}`

**前端动作**:
1. **状态更新**：从该群的"成员列表"中移除该用户。
2. (可选) 在聊天窗口插入一条系统消息："用户 [999] 退出了群聊"。

### 4. 新成员加入 (joinGroup)

**触发时机**: 接受邀请入群 或 管理员同意加群申请。
**数据示例**: `{"userId": 999, "groupId": "1001"}`

**前端动作**:
*   **如果是你加入了新群 (`userId == currentUser.id`)**:
    1. **API调用**：立即调用 `GET /api/v1/group/:id` 获取该群的详细信息。
    2. **状态更新**：将新群添加到本地群组列表的最上方。
*   **如果是别人加入**:
    1. **状态更新**：将该用户添加到"成员列表"中。
    2. (可选) 在聊天窗口插入一条系统消息："欢迎用户 [999] 加入群聊"。

---

## 心跳机制

### 心跳配置

| 参数 | 值 | 说明 |
|------|-----|------|
| PingInterval | 30秒 | 客户端发送心跳间隔 |
| PongTimeout | 60秒 | 服务端等待 pong 超时时间 |

### 心跳流程

```
客户端每30秒发送一次 ping
    ↓
{"type": "ping"}
    ↓
服务端收到后立即响应 pong
    ↓
{"type": "pong", "data": {"timestamp": 1736683200}}
    ↓
如果60秒内未收到客户端的 ping
    ↓
服务端主动断开连接
```

### Ping 消息

**客户端发送**:
```json
{
  "type": "ping"
}
```

**服务端响应**:
```json
{
  "type": "pong",
  "data": {
    "timestamp": 1736683200
  }
}
```

### 重连策略

**建议的重连策略**:

1. **指数退避**:
   - 第1次重连：立即
   - 第2次重连：1秒后
   - 第3次重连：2秒后
   - 第4次重连：4秒后
   - 第5次及以后：8秒后

2. **网络变化监听**:
   - 监听网络状态变化
   - 网络恢复后立即重连

3. **用户主动重连**:
   - 提供"重新连接"按钮

---

## 离线消息推送

### 推送时机

用户建立 WebSocket 连接后，服务端自动推送离线消息。

### 推送规则

**私聊离线消息**:
- 从 Redis 离线队列获取前 20 条
- 如果总数 > 20，设置 `hasMore=true`
- 剩余消息通过 HTTP 接口拉取

**群聊离线消息**:
- **已支持推送**: 同样推送最近 20 条离线群消息
- `offline_messages` 类型将包含 `messageType: "group"`

### 消息去重

客户端应根据 `msgId` 去重：

```
收到消息时检查本地是否已存在
if (messageExists(msgId)) {
    忽略
} else {
    显示消息
}
```

---

## 错误处理

### 连接错误

| 错误码 | 说明 | 处理方式 |
|-------|------|---------|
| 1008 | Token 无效或过期 | 刷新 Token 后重连 |
| 1000 | 正常关闭 | 正常，无需特殊处理 |
| 1001 | 服务端主动断开 | 重连 |
| 1006 | 连接异常 | 检查网络，重连 |

### 消息错误

**发送失败**:
```json
{
  "type": "error",
  "data": {
    "code": 30004,
    "message": "对方不是好友，无法发送消息"
  }
}
```

**常见错误码**:
| code | 说明 |
|------|------|
| 30001 | 参数错误 |
| 30004 | 对方不是好友 |
| 30006 | 不是群成员 |
| 30007 | 被禁言 |
| 30008 | @全体成员需要管理员权限 |

---

## 常见问题

### Q1: 如何判断连接是否成功？

**A**: 
1. WebSocket 连接建立（`onopen` 事件触发）
2. 收到 `type="connected"` 消息

### Q2: 收不到消息怎么办？

**A**: 检查以下几点：
1. WebSocket 连接是否正常
2. Token 是否过期
3. 是否正确处理消息类型
4. 检查浏览器控制台错误

### Q3: 如何优雅地断开连接？

**A**: 
```
发送关闭帧
websocket.close(1000, "正常关闭");
```

### Q4: 离线消息最多推送多少条？

**A**: 
- WebSocket 连接时自动推送前 **20 条**
- 剩余离线消息需调用 HTTP 接口拉取

### Q5: 心跳超时会怎样？

**A**: 
- 60秒内未收到客户端 ping
- 服务端主动断开连接（关闭码 1001）
- 客户端需要重连

### Q6: 如何实现消息重发？

**A**: 
1. 客户端生成唯一 `msgId`
2. 发送消息时保存到本地
3. 如果发送失败，使用相同 `msgId` 重发
4. 服务端根据 `msgId` 去重

### Q7: 同一账号多设备登录怎么办？

**A**: 
- 支持多设备同时在线
- 每个设备独立 WebSocket 连接
- 消息会推送给所有在线设备

### Q8: WebSocket 断线后未读消息会丢失吗？

**A**: 
- 不会。消息存储在数据库中
- 重连后自动推送离线消息
- 可以通过 HTTP 接口拉取历史消息

---

## 完整示例

### 连接流程

```
1. 获取 Token
   POST /api/v1/auth/login
   → 返回 accessToken

2. 建立 WebSocket 连接
   ws://localhost:10300/ws?token=<accessToken>
   
3. 收到连接成功消息
   {"type": "connected", "data": {...}}
   
4. 收到离线消息
   {"type": "offline_messages", "data": {...}}
   
5. 开始心跳（每30秒）
   → {"type": "ping"}
   ← {"type": "pong", "data": {...}}
   
6. 发送/接收消息
   → {"type": "chat", "data": {...}}
   ← {"type": "chat", "data": {...}}
```

### 消息收发流程

```
发送私聊消息:
→ {"type": "chat", "data": {toUserId: 1002, content: "你好"}}
← {"type": "chat", "data": {id: 123, msgId: "...", ...}}

接收私聊消息:
← {"type": "chat", "data": {fromUserId: 1003, content: "在吗"}}

发送群聊消息:
→ {"type": "group_chat", "data": {groupId: "g1", content: "大家好"}}
← {"type": "group_chat", "data": {seq: 1250, ...}}

接收群聊消息:
← {"type": "group_chat", "data": {groupId: "g1", content: "你好"}}
```

---

**文档维护**: Skylm  
**最后更新**: 2026-01-15  
**相关文档**: [WebSocket 架构设计](./ARCHITECTURE.md)
