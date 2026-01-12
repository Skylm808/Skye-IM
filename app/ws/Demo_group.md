# Beaver IM 群聊功能与 WebSocket 结合分析

本文档分析了 Beaver IM 项目中群聊功能的实现原理，以及它是如何与 WebSocket 模块结合以实现实时通信的。

## 1. 架构概览

该项目采用微服务架构（基于 go-zero），将业务逻辑拆分为多个独立的服务。在群聊功能中，主要涉及以下三个核心服务：

*   **WS 服务 (`app/ws`)**: 作为**网关 (Gateway)**，负责维护 WebSocket 长连接、消息路由（上行）和消息推送（下行）。它不处理复杂的业务逻辑。
*   **Chat 服务 (`app/chat`)**: 作为**消息逻辑核心**，负责消息的存储、会话管理（Conversation）、消息序列号生成以及消息的扩散（Fan-out）。
*   **Group 服务 (`app/group`)**: 作为**群组业务核心**，负责群组的创建、成员管理（加群/退群）、群信息修改等。

## 2. WebSocket 模块结合方式

WebSocket 模块 (`app/ws/ws_api`) 是实时通信的基石，它通过以下方式与其他模块结合：

### 2.1 连接管理
*   **连接升级**: `chatwebsockethandler.go` 处理 HTTP Upgrade 请求。
*   **连接维护**: `ChatWebsocketLogic` 维护本地内存中的连接映射 (`UserOnlineWsMap`)，并将在线状态同步到 Redis (`redis_ws`)。这使得系统知道用户连接在哪个 WS 节点上。

### 2.2 上行消息 (Client -> Server)
客户端通过 WebSocket 发送的消息（如“发送群消息”）会被 WS 服务接收并转发给业务服务。

*   **入口**: `ws_api/internal/logic/websocket/enter.go` 中的 `HandleWebSocketMessages` 是总入口。
*   **路由**: 根据消息的 `Command` 字段进行分发。
    *   `CHAT_MESSAGE` -> `chat_message.HandleChatMessageTypes`
*   **转发**:
    *   对于群消息 (`GroupMessageSend`)，WS 服务**不直接处理**，而是将其转换为 RPC 请求，调用 **Chat 服务** 的 `ChatRpc.SendMsg` 接口。
    *   代码路径: `app/ws/ws_api/internal/logic/websocket/chat_message/group_message_send.go`。

### 2.3 下行消息 (Server -> Client)
当业务服务（如 Chat 或 Group）需要推送消息给客户端时，它们不直接持有 WebSocket 连接，而是通过 HTTP 回调 WS 服务的代理接口。

*   **代理接口**: WS 服务暴露了一个内部 HTTP 接口 `/api/ws/proxySendMsg`。
*   **处理逻辑**: `proxysendmsglogic.go` 接收请求，解析目标 UserID，在本地连接映射中找到对应的 `websocket.Conn`，将消息写入连接。
*   **调用方**: Chat 服务或 Group 服务使用 `common/ajax/enter.go` 中的 `SendMessageToWs` 工具函数来调用此接口。

## 3. 群聊功能核心流程详解

### 3.1 发送群聊消息流程

这是一个典型的 **Logic 与 Gateway 分离** 的设计。

1.  **用户发送**: 客户端通过 WebSocket 发送 `GROUP_MESSAGE_SEND` 消息。
2.  **WS 转发**:
    *   WS 服务收到消息，识别为群消息。
    *   构造 `chat_rpc.SendMsgReq`。
    *   调用 `svcCtx.ChatRpc.SendMsg` (gRPC)。
3.  **Chat 服务处理** (`app/chat/chat_rpc/internal/logic/sendmsglogic.go`):
    *   **入库**: 将消息保存到 MySQL 的 `messages` 表。
    *   **更新会话**: 更新 `conversations` 表（最新消息预览、版本号）。
    *   **读扩散优化**: 系统维护了 `user_conversations` 表，记录了用户参与的所有会话。Chat 服务通过此表查找该群聊会话下的所有成员 (`recipientIds`)。
    *   **消息分发 (Fan-out)**: 遍历所有成员，调用 `ajax.SendMessageToWs`。
4.  **WS 推送**:
    *   WS 服务收到 `proxySendMsg` 请求。
    *   将消息推送到对应用户的客户端。

### 3.2 创建群组流程

创建群组涉及 `Group` 服务和 `Chat` 服务的协同。

1.  **创建请求**: 客户端调用 Group API (`/api/group/create`)。
2.  **Group 服务处理** (`app/group/group_api/internal/logic/groupcreatelogic.go`):
    *   **业务落地**: 在 `groups` 和 `group_members` 表中创建记录。
    *   **初始化会话**: 异步调用 `ChatRpc.InitializeConversation`。这一步在 Chat 服务中建立了 `conversationId` (如 `group_UUID`) 与成员的映射，为后续发消息做准备。
    *   **发送系统通知**: 异步调用 `ChatRpc.SendNotificationMessage`，生成一条“XXX 创建了群聊”的系统消息（走标准消息推送流程）。
    *   **同步数据**: 异步调用 `ajax.SendMessageToWs`，发送 `GROUP_OPERATION` 指令，通知客户端拉取新的群组列表和成员列表。

## 4. 关键代码文件索引

| 模块 | 功能 | 文件路径 |
| :--- | :--- | :--- |
| **WS** | WS 消息入口 | `app/ws/ws_api/internal/logic/websocket/enter.go` |
| **WS** | 群消息转发 | `app/ws/ws_api/internal/logic/websocket/chat_message/group_message_send.go` |
| **WS** | 消息推送代理 | `app/ws/ws_api/internal/logic/proxysendmsglogic.go` |
| **WS** | 推送工具类 | `common/ajax/enter.go` |
| **Chat** | 发送消息逻辑 | `app/chat/chat_rpc/internal/logic/sendmsglogic.go` |
| **Group** | 创建群组逻辑 | `app/group/group_api/internal/logic/groupcreatelogic.go` |

## 5. 总结

该项目的群聊实现具有以下特点：

1.  **职责单一**: WS 层只管连接，业务层（Chat/Group）只管逻辑。
2.  **双向通信解耦**:
    *   上行通过 RPC (WS -> Logic)。
    *   下行通过 HTTP 回调 (Logic -> WS)。
3.  **会话抽象**: 群聊被抽象为一种 `Conversation`，消息投递依赖于会话成员关系表 (`user_conversations`)，而不是直接查询群成员表，这有助于统一处理私聊和群聊的消息逻辑。
