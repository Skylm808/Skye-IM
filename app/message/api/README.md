# Message API 模块

## 概述
Message API 服务提供消息相关的 HTTP 接口，供前端调用。

## 服务端口
- **HTTP 服务**: `http://localhost:10400`

## 启动方式
```powershell
cd app/message/api
go run message.go
```

## API 接口

### 1. 获取历史消息
```
GET /api/v1/message/history?peerId=2&lastMsgId=0&limit=20
```
**Headers:** `Authorization: Bearer <token>`

**响应：**
```json
{
  "list": [
    {
      "id": 1,
      "msgId": "uuid-xxx",
      "fromUserId": 1,
      "toUserId": 2,
      "content": "你好",
      "contentType": 1,
      "status": 0,
      "createdAt": 1704441600
    }
  ],
  "hasMore": false
}
```

### 2. 获取未读消息数
```
GET /api/v1/message/unread/count?peerId=2
```
**响应：**
```json
{
  "count": 5
}
```

### 3. 标记消息为已读
```
POST /api/v1/message/read
Content-Type: application/json
```
**请求体：**
```json
{
  "peerId": 2,
  "msgIds": ["uuid-1", "uuid-2"]
}
```
**响应：**
```json
{
  "count": 2
}
```

### 4. 获取会话列表
```
GET /api/v1/message/conversations
```
**响应：**
```json
{
  "list": []
}
```
