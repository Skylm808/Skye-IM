# Message RPC 模块

## 概述
Message RPC 服务提供消息的持久化存储和查询功能。

## 服务端口
- **RPC 服务**: `127.0.0.1:9300`
- **Etcd Key**: `message.rpc`

## 启动方式

### 1. 创建消息表
```sql
source c:/Users/Tianlinmao/Desktop/GoLand/SkyeIM/app/message/im_message.sql
```

### 2. 启动服务
```powershell
cd app/message/rpc
go run message.go
```

## RPC 接口

| 方法 | 说明 |
|------|------|
| `SendMessage` | 发送消息（存储到数据库） |
| `GetMessageList` | 获取历史消息列表（分页） |
| `MarkAsRead` | 标记消息为已读 |
| `GetUnreadCount` | 获取未读消息数量 |
| `GetUnreadMessages` | 获取与某用户的未读消息 |

## 消息类型

| contentType | 说明 |
|-------------|------|
| 1 | 文字消息 |
| 2 | 图片消息 |
| 3 | 文件消息 |
| 4 | 语音消息 |

## 消息状态

| status | 说明 |
|--------|------|
| 0 | 未读 |
| 1 | 已读 |
| 2 | 撤回 |
