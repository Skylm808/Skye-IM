# Message 模块前端对接文档

## 📋 目录

- [概述](#概述)
- [私聊消息接口](#私聊消息接口)
- [群聊消息接口](#群聊消息接口)
- [会话管理接口](#会话管理接口)
- [消息搜索接口](#消息搜索接口)
- [数据字段说明](#数据字段说明)
- [错误码说明](#错误码说明)

---

## 概述

### Base URL

```
http://localhost:8080
```

### 通用请求头

```
Authorization: Bearer <access_token>
Content-Type: application/json
```

### 接口总览

| 模块 | 接口数 | 说明 |
|------|-------|------|
| 私聊消息 | 5个 | 发送、历史、离线同步、未读、已读 |
| 群聊消息 | 4个 | 发送、历史、离线同步、已读上报 |
| 会话管理 | 1个 | 最近会话列表 |
| 消息搜索 | 2个 | 模糊搜索、@我的消息 |

**共计**: 12个API接口

**注意**: 发送消息主要通过 WebSocket，HTTP 接口为可选备用方案。

---

## 私聊消息接口

### 1. 发送私聊消息（HTTP 可选）

**场景**: HTTP方式发送私聊消息（建议使用WebSocket）

**端点**: `POST /api/v1/message/send`

**请求体**:
```json
{
  "toUserId": 1002,
  "content": "你好，在吗？",
  "contentType": 1
}
```

**字段说明**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|-----|------|
| toUserId | int64 | 是 | 接收者用户ID |
| content | string | 是 | 消息内容 |
| contentType | int32 | 是 | 内容类型：1-文本 2-图片 3-文件 4-语音 |

**成功响应** (200):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 12345,
    "msgId": "msg_20260113_12345",
    "createdAt": 1736683200
  }
}
```

**注意事项**:
- 建议优先使用 WebSocket 发送消息
- HTTP 接口适用于 WebSocket 断线时的降级方案

---

### 2. 获取私聊历史消息

**场景**: 查看与某用户的聊天记录

**端点**: `GET /api/v1/message/history`

**查询参数**:
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|-----|-------|------|
| peerId | int64 | 是 | - | 对方用户ID |
| lastMsgId | int64 | 否 | 0 | 最后一条消息ID（用于分页） |
| limit | int32 | 否 | 20 | 获取条数 |

**请求示例**:
```
GET /api/v1/message/history?peerId=1002&lastMsgId=12340&limit=20
```

**成功响应** (200):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 12345,
        "msgId": "msg_20260113_12345",
        "fromUserId": 1001,
        "toUserId": 1002,
        "chatType": 1,
        "content": "你好，在吗？",
        "contentType": 1,
        "status": 1,
        "createdAt": 1736683200
      }
    ],
    "hasMore": true
  }
}
```

**分页说明**:
- 首次加载：不传 `lastMsgId`
- 加载更多：传上一页最后一条消息的 `id` 作为 `lastMsgId`
- `hasMore=true` 表示还有更多历史消息

---

### 3. 私聊离线同步

**场景**: WebSocket 连接成功后，拉取剩余离线消息

**端点**: `GET /api/v1/message/offline`

**查询参数**:
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|-----|-------|------|
| skip | int32 | 否 | 0 | 跳过前N条（已通过WS推送的） |
| limit | int32 | 否 | 100 | 每次拉取条数 |

**请求示例**:
```
GET /api/v1/message/offline?skip=20&limit=100
```

**成功响应** (200):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 12340,
        "msgId": "msg_20260113_12340",
        "fromUserId": 1003,
        "toUserId": 1001,
        "chatType": 1,
        "content": "晚上一起吃饭吗？",
        "contentType": 1,
        "status": 0,
        "createdAt": 1736683100
      }
    ],
    "hasMore": false,
    "total": 20
  }
}
```

**使用场景**:
1. WebSocket 连接成功
2. WebSocket 自动推送前 20 条离线消息
3. 如果 total > 20，调用此接口拉取剩余消息

---

### 4. 获取私聊未读数

**场景**: 获取未读消息数量

**端点**: `GET /api/v1/message/unread/count`

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|-----|------|
| peerId | int64 | 否 | 对方用户ID，为空则获取所有未读数 |

**请求示例**:
```
# 获取与某用户的未读数
GET /api/v1/message/unread/count?peerId=1002

# 获取所有未读数
GET /api/v1/message/unread/count
```

**成功响应** (200):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "count": 5
  }
}
```

---

### 5. 标记私聊消息为已读

**场景**: 用户阅读消息后上报已读状态

**端点**: `POST /api/v1/message/read`

**请求体**:
```json
{
  "peerId": 1002,
  "msgIds": ["msg_20260113_12340", "msg_20260113_12341"]
}
```

**字段说明**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|-----|------|
| peerId | int64 | 是 | 对方用户ID |
| msgIds | []string | 否 | 消息ID列表，为空则标记全部 |

**成功响应** (200):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "count": 2
  }
}
```

**注意事项**:
- 建议批量标记已读，避免频繁请求
- `msgIds` 为空时，标记与该用户的所有未读消息

---

## 群聊消息接口

### 1. 发送群聊消息（HTTP 可选）

**场景**: HTTP方式发送群聊消息（建议使用WebSocket）

**端点**: `POST /api/v1/message/group/send`

**请求体**:
```json
{
  "groupId": "g_20260113_001",
  "content": "@张三 @李四 明天开会",
  "contentType": 1,
  "atUserIds": [1002, 1003]
}
```

**字段说明**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|-----|------|
| groupId | string | 是 | 群组ID |
| content | string | 是 | 消息内容 |
| contentType | int32 | 是 | 内容类型：1-文本 2-图片 3-文件 4-语音 |
| atUserIds | []int64 | 否 | 被@的用户ID列表，-1表示@全体成员 |

**成功响应** (200):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 12350,
    "msgId": "msg_20260113_12350",
    "createdAt": 1736683300,
    "seq": 1250
  }
}
```

**注意事项**:
- `seq` 为群消息序列号，用于离线同步和已读进度
- `atUserIds` 为 `[-1]` 表示@全体成员（需要群主或管理员权限）

---

### 2. 获取群聊历史消息

**场景**: 查看群聊的聊天记录

**端点**: `GET /api/v1/message/group/history`

**查询参数**:
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|-----|-------|------|
| groupId | string | 是 | - | 群组ID |
| lastMsgId | int64 | 否 | 0 | 最后一条消息ID（用于分页） |
| limit | int32 | 否 | 20 | 获取条数 |

**请求示例**:
```
GET /api/v1/message/group/history?groupId=g_20260113_001&limit=20
```

**成功响应** (200):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 12350,
        "msgId": "msg_20260113_12350",
        "fromUserId": 1001,
        "chatType": 2,
        "groupId": "g_20260113_001",
        "content": "@张三 @李四 明天开会",
        "contentType": 1,
        "status": 0,
        "createdAt": 1736683300,
        "seq": 1250,
        "atUserIds": [1002, 1003]
      }
    ],
    "hasMore": true
  }
}
```

---

### 3. 群聊离线同步（按 Seq）

**场景**: WebSocket 连接后，拉取群聊离线消息

**端点**: `GET /api/v1/message/group/sync`

**查询参数**:
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|-----|-------|------|
| groupId | string | 是 | - | 群组ID |
| seq | uint64 | 否 | 0 | 起始Seq（不包含），拉取 > seq 的消息 |
| limit | int32 | 否 | 200 | 获取条数（后端会限制上限） |

**请求示例**:
```
GET /api/v1/message/group/sync?groupId=g_20260113_001&seq=1200&limit=200
```

**成功响应** (200):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 12345,
        "msgId": "msg_20260113_12345",
        "fromUserId": 1002,
        "chatType": 2,
        "groupId": "g_20260113_001",
        "content": "大家好",
        "contentType": 1,
        "status": 0,
        "createdAt": 1736683200,
        "seq": 1201,
        "atUserIds": []
      }
    ]
  }
}
```

**使用场景**:
1. 用户上次已读到 `seq=1200`
2. 离线后群里有新消息 `seq=1201~1250`
3. 上线后调用此接口拉取 `seq > 1200` 的消息

---

### 4. 群聊已读上报

**场景**: 用户阅读群聊消息后上报已读进度

**端点**: `POST /api/v1/message/group/read`

**请求体**:
```json
{
  "groupId": "g_20260113_001",
  "readSeq": 1250
}
```

**字段说明**:
| 字段 | 类型 | 必填 | 说明 |
|------|------|-----|------|
| groupId | string | 是 | 群组ID |
| readSeq | uint64 | 是 | 已读到的消息Seq |

**成功响应** (200):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "success": true
  }
}
```

**注意事项**:
- 定期上报已读进度（如每 30 秒或退出聊天时）
- 用于计算未读数：未读数 = 最新Seq - readSeq

---

## 会话管理接口

### 1. 获取会话列表

**场景**: 获取最近联系人列表

**端点**: `GET /api/v1/message/conversations`

**无查询参数**

**成功响应** (200):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "peerId": 1002,
        "lastMessage": {
          "id": 12345,
          "msgId": "msg_20260113_12345",
          "fromUserId": 1002,
          "toUserId": 1001,
          "chatType": 1,
          "content": "明天见",
          "contentType": 1,
          "status": 0,
          "createdAt": 1736683200
        },
        "unreadCount": 3
      },
      {
        "peerId": 1003,
        "lastMessage": {
          "id": 12340,
          "msgId": "msg_20260113_12340",
          "fromUserId": 1001,
          "toUserId": 1003,
          "chatType": 1,
          "content": "好的",
          "contentType": 1,
          "status": 1,
          "createdAt": 1736683100
        },
        "unreadCount": 0
      }
    ]
  }
}
```

**注意事项**:
- 按最后一条消息时间倒序排列
- 包含未读消息数

---

## 消息搜索接口

### 1. 模糊搜索聊天记录

**场景**: 搜索消息内容

**端点**: `GET /api/v1/message/search`

**查询参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|-----|------|
| keyword | string | 是 | 搜索关键词 |

**请求示例**:
```
GET /api/v1/message/search?keyword=开会
```

**成功响应** (200):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 12350,
        "msgId": "msg_20260113_12350",
        "fromUserId": 1001,
        "toUserId": 1002,
        "chatType": 1,
        "content": "明天开会记得带笔记本",
        "contentType": 1,
        "status": 1,
        "createdAt": 1736683300
      }
    ]
  }
}
```

**注意事项**:
- 支持模糊匹配消息内容
- 只返回用户参与的会话消息

---

### 2. 获取@我的群消息

**场景**: 查看群聊中@我的消息

**端点**: `GET /api/v1/message/at-me`

**查询参数**:
| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|-----|-------|------|
| groupId | string | 否 | - | 群组ID，为空则查询所有群 |
| lastMsgId | int64 | 否 | 0 | 最后一条消息ID（用于分页） |
| limit | int32 | 否 | 20 | 获取条数 |

**请求示例**:
```
# 查询所有群的@我的消息
GET /api/v1/message/at-me?limit=20

# 查询指定群的@我的消息
GET /api/v1/message/at-me?groupId=g_20260113_001&limit=20
```

**成功响应** (200):
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "list": [
      {
        "id": 12350,
        "msgId": "msg_20260113_12350",
        "fromUserId": 1001,
        "chatType": 2,
        "groupId": "g_20260113_001",
        "content": "@张三 明天开会",
        "contentType": 1,
        "status": 0,
        "createdAt": 1736683300,
        "seq": 1250,
        "atUserIds": [1002]
      }
    ],
    "hasMore": false
  }
}
```

**注意事项**:
- 包含@我和@全体成员的消息
- 可用于"@提醒"功能

---

## 数据字段说明

### MessageInfo 字段

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int64 | 消息ID（数据库主键） |
| msgId | string | 消息唯一标识 |
| fromUserId | int64 | 发送者ID |
| toUserId | int64 | 接收者ID（私聊时有效） |
| chatType | int32 | 聊天类型：1-私聊 2-群聊 |
| groupId | string | 群组ID（群聊时使用） |
| content | string | 消息内容 |
| contentType | int32 | 内容类型：1-文本 2-图片 3-文件 4-语音 |
| status | int32 | 消息状态：0-未读 1-已读 2-撤回 |
| createdAt | int64 | 创建时间（Unix时间戳，秒） |
| seq | uint64 | 群消息序列号（仅群聊） |
| atUserIds | []int64 | 被@的用户ID列表，-1表示@全体 |

### 消息状态说明

| status | 说明 |
|--------|------|
| 0 | 未读/未处理 |
| 1 | 已读 |
| 2 | 已撤回 |

### 内容类型说明

| contentType | 说明 |
|-------------|------|
| 1 | 文本消息 |
| 2 | 图片消息 |
| 3 | 文件消息 |
| 4 | 语音消息 |

---

## 错误码说明

| 错误码 | 说明 |
|-------|------|
| 0 | 成功 |
| 30001 | 参数错误 |
| 30002 | 消息不存在 |
| 30003 | 无权访问该消息 |
| 30004 | 对方不是好友，无法发送消息 |
| 30005 | 群组不存在或已解散 |
| 30006 | 不是群成员，无法发送消息 |
| 30007 | 被禁言，无法发送消息 |
| 30008 | @全体成员需要管理员权限 |

---

## 消息推送机制

### WebSocket vs HTTP

| 场景 | 推荐方式 | 说明 |
|------|---------|------|
| 发送消息 | WebSocket | 实时性好，服务器压力小 |
| 接收消息 | WebSocket | 实时推送 |
| 历史消息 | HTTP API | 按需拉取 |
| 离线消息 | WebSocket + HTTP | WS连接时推送前20条，剩余HTTP拉取 |

### 完整流程

```
1. 用户登录
   ↓
2. 建立 WebSocket 连接
   ws://localhost:10300/ws?token=xxx
   ↓
3. WebSocket 自动推送离线消息（前20条）
   ↓
4. 如果离线消息 > 20条，HTTP拉取剩余
   GET /api/v1/message/offline?skip=20&limit=100
   ↓
5. 实时收发消息
   通过 WebSocket
   ↓
6. 查看历史记录
   GET /api/v1/message/history
```

---

## 常见问题

### Q1: 消息列表如何分页？

**A**: 使用 `lastMsgId` 分页：

```
# 第一页
GET /api/v1/message/history?peerId=1002&limit=20

# 第二页（传上一页最后一条消息的id）
GET /api/v1/message/history?peerId=1002&lastMsgId=12320&limit=20
```

### Q2: 如何判断是否还有更多消息？

**A**: 检查响应中的 `hasMore` 字段：
- `hasMore: true` - 还有更多消息
- `hasMore: false` - 已加载全部

### Q3: 群聊未读数如何计算？

**A**: 
```
未读数 = 群最新消息Seq - 用户readSeq
```
前端需要维护每个群的 `readSeq`，定期上报给后端。

### Q4: WebSocket 断线后如何同步消息？

**A**: 
1. 重连 WebSocket
2. 私聊：调用 `/offline` 接口拉取离线消息
3. 群聊：调用 `/group/sync` 接口，传上次的 `readSeq`

### Q5: 如何实现消息撤回？

**A**: 当前未实现，如需实现：
1. 添加撤回接口：`POST /api/v1/message/recall`
2. 更新消息 `status=2`
3. WebSocket 推送撤回通知给对方

---

**文档维护**: Skylm  
**最后更新**: 2026-01-13  
**相关文档**: [Message 架构设计](../ARCHITECTURE.md)
