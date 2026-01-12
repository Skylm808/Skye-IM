# message 模块（消息服务）

本目录负责 **IM 消息的存储与查询**，同时提供 HTTP API（给前端拉历史、上报已读、离线同步等）与 RPC（给 WS/其它服务调用）。

## 1. 职责边界

- **消息落库**：私聊/群聊消息写入 `im_message` 表（`app/message/model`）。
- **消息查询**：历史消息分页、未读消息/未读数、按 seq 拉取群聊增量。
- **已读处理**：
  - 私聊：`message-rpc.MarkAsRead` 更新消息状态（面向点对点）。
  - 群聊：已读进度 `read_seq` 归属在 `group` 模块（群成员维度），message 侧只提供“按 seq 拉取增量”。
- **序列号（群聊 seq）生成**：群聊消息的 `seq` 通过 Redis 自增生成（见 `app/message/rpc/internal/logic/sendgroupmessagelogic.go`）。

## 2. 数据模型（核心字段）

`im_message` 同时承载私聊与群聊：

- `chat_type`：`1=私聊`，`2=群聊`
- `to_user_id`：私聊目标用户（群聊通常为 0）
- `group_id`：群聊目标群（私聊为空）
- `seq`：**仅群聊使用**，用于离线同步与已读进度（按群维度单调递增）

对应 SQL：`app/message/im_message.sql`。

## 3. 对外接口

### 3.1 HTTP（message-api）

入口：`app/message/api/message.api`

常用：
- `GET /api/v1/message/history`：私聊历史（分页）
- `GET /api/v1/message/group/history`：群聊历史（分页）
- `GET /api/v1/message/group/sync`：群聊离线同步（按 `seq` 拉取 `> seq`）
- `POST /api/v1/message/read`：私聊已读上报
- `POST /api/v1/message/group/read`：群聊已读上报（更新成员 `read_seq`，实际落到 `group-rpc.UpdateGroupReadSeq`）

说明：
- “历史消息/离线同步/已读上报” 用 HTTP 做是合理的：便于分页、可重试、可观测。
- `POST /send`、`POST /group/send` 在设计上是 **可选兜底/调试**；主链路建议走 WS（见 `app/ws`）。

### 3.2 RPC（message-rpc）

入口：`app/message/rpc/message.proto`

与群聊最相关的方法：
- `SendGroupMessage`：写入群聊消息，并生成 `seq`
- `GetGroupMessageList`：群聊历史分页（按消息 id）
- `GetGroupMessagesBySeq`：群聊增量（按 `seq`）

## 4. 群聊“离线同步”的核心点（为什么要 seq）

群聊的已读/未读判断，本质是：

- 每个成员在每个群里维护一个 `read_seq`（存储在 `im_group_member.read_seq`）。
- 群消息写入时生成严格递增 `seq`。
- 用户上线或进入群聊时：
  - 用 `read_seq` 去拉 `seq > read_seq` 的消息（message-rpc / message-api 都支持）。
  - 上报新的 `read_seq`（通常是本地已展示/已读到的最大 seq）。

这套机制相比“按 msg_id / created_at”更稳定：不怕乱序、不依赖时钟、增量同步简单。

## 5. 你后续面试加分的增强点（建议）

- 会话列表（Conversations）：按用户维度维护最近会话、最后一条消息、未读数（可落库或 Redis）。
- 幂等与去重：`msg_id` 做唯一键（或业务侧幂等表），避免客户端重试导致重复消息。
- 消息状态机：`sent/delivered/read` 的一致定义与落库策略（尤其群聊的 delivered/read 聚合）。
- 大群优化：批量推送/限流/分片；推送链路解耦（Kafka）。

