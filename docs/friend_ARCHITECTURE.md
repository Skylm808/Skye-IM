# Friend 微服务架构与流程讲解

## 一、服务概述

Friend 服务是 SkyeIM 即时通讯系统的**好友关系管理微服务**，负责好友申请、好友列表、黑名单等核心社交功能。

### 技术栈

| 组件 | 技术选型 | 用途 |
|------|---------|------|
| 框架 | go-zero | 微服务框架 |
| 数据库 | MySQL | 好友关系持久化 |
| 缓存 | Redis | Model 缓存 |
| 通信 | gRPC + HTTP | RPC 内部调用 + HTTP 对外接口 |

---

## 二、目录结构

```
app/friend/
├── api/                         # HTTP API 层
│   ├── friend.api               # API 定义文件
│   ├── etc/
│   │   └── friend-api.yaml     # API 服务配置
│   └── internal/
│       ├── config/             # 配置结构体
│       ├── handler/            # HTTP 处理器（自动生成）
│       │   ├── blacklist/      # 黑名单处理器
│       │   ├── friend/         # 好友关系处理器
│       │   └── request/        # 好友申请处理器
│       ├── logic/              # 业务逻辑层
│       │   ├── blacklist/      # 黑名单逻辑
│       │   ├── friend/         # 好友关系逻辑
│       │   └── request/        # 好友申请逻辑
│       ├── svc/
│       │   └── serviceContext.go  # 服务上下文
│       └── types/
│           └── types.go        # 请求/响应类型
├── rpc/                         # gRPC 服务层
│   ├── etc/
│   │   └── friend.yaml         # RPC 服务配置
│   ├── internal/
│   │   ├── logic/              # RPC 业务逻辑
│   │   └── svc/                # RPC 服务上下文
│   ├── friend.go               # RPC 服务入口
│   └── friend.proto            # Protobuf 定义（未使用 proto，直接通过 API 调用）
├── model/                       # 数据模型层
│   ├── im_friend.sql            # 好友表 DDL
│   ├── im_friend_request.sql    # 申请表 DDL
│   ├── imfriendmodel.go         # 好友 Model 接口
│   ├── imfriendmodel_gen.go     # 好友 Model 实现（带缓存）
│   ├── imfriendrequestmodel.go  # 申请 Model 接口
│   └── imfriendrequestmodel_gen.go
└── FRIEND_API文档.md             # 前端对接文档
```

---

## 三、数据库设计

### 3.1 好友关系表 (im_friend)

```sql
CREATE TABLE `im_friend` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned NOT NULL COMMENT '用户ID',
  `friend_id` bigint unsigned NOT NULL COMMENT '好友ID',
  `remark` varchar(50) DEFAULT '' COMMENT '好友备注',
  `status` tinyint NOT NULL DEFAULT 1 COMMENT '1-正常 2-拉黑',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_friend` (`user_id`, `friend_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='好友关系表';
```

**设计要点**：
- 🔑 **唯一索引** `uk_user_friend`：防止重复添加好友
- 🔗 **双向关系**：A 添加 B 为好友，需要插入两条记录
  - (user_id=A, friend_id=B)
  - (user_id=B, friend_id=A)
- 📝 **备注字段**：每个用户可为好友设置独立备注
- 🚫 **状态字段**：1-正常，2-拉黑

### 3.2 好友申请表 (im_friend_request)

```sql
CREATE TABLE `im_friend_request` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `from_user_id` bigint unsigned NOT NULL COMMENT '发起人ID',
  `to_user_id` bigint unsigned NOT NULL COMMENT '接收人ID',
  `message` varchar(200) DEFAULT '' COMMENT '验证消息',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0-待处理 1-已同意 2-已拒绝',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_to_user` (`to_user_id`, `status`),
  KEY `idx_from_user` (`from_user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='好友申请表';
```

**设计要点**：
- 🔍 **索引优化**：
  - `idx_to_user`: 查询收到的申请（常用场景）
  - `idx_from_user`: 查询发出的申请
- 📨 **单向申请**：一条记录代表一次申请
- 🔄 **状态流转**：0(待处理) → 1(已同意) / 2(已拒绝)

---

## 四、核心流程

### 4.1 发送好友申请流程

```
┌─────────┐     POST /api/v1/friend/request      ┌─────────────┐
│  前端   │ ───────────────────────────────────► │  Handler    │
└─────────┘  {toUserId, message}                  └──────┬──────┘
                                                          │
                                                          ▼
                                                   ┌─────────────┐
                                                   │   Logic     │
                                                   │ AddFriend   │
                                                   │  Request    │
                                                   └──────┬──────┘
                                                          │
                ┌─────────────────────────────────────────┼─────────────┐
                │                                         │             │
                ▼                                         ▼             ▼
         ┌─────────────┐                         ┌─────────────┐ ┌─────────────┐
         │ 检查是否     │                         │ 检查是否     │ │ 检查是否     │
         │ 已是好友    │                         │ 重复申请    │ │ 申请自己    │
         └──────┬──────┘                         └──────┬──────┘ └──────┬──────┘
                │                                       │               │
                │ 查询 im_friend                        │ 查询申请表    │ userId == friendId?
                ▼                                       ▼               ▼
         ┌─────────────┐                         ┌─────────────┐ ┌─────────────┐
         │ 已是好友    │                         │ 待处理申请  │ │ 返回错误    │
         │ 返回错误    │                         │ 返回错误    │ └─────────────┘
         └─────────────┘                         └─────────────┘
                
                                                          │
                                                          ▼
                                                   ┌─────────────┐
                                                   │ 插入申请记录│
                                                   │ im_friend_  │
                                                   │   request   │
                                                   └──────┬──────┘
                                                          │
                                                          ▼
                                                   ┌─────────────┐
                                                   │ 返回申请ID  │
                                                   └─────────────┘
```

**关键检查**：
1. ❌ 不能向自己发送申请
2. ❌ 已是好友则不能申请
3. ❌ 已有待处理申请则不能重复申请

---

### 4.2 处理好友申请流程（同意）

```
┌─────────┐     PUT /api/v1/friend/request/123    ┌─────────────┐
│  前端   │ ───────────────────────────────────► │  Handler    │
└─────────┘  {action: 1}                          └──────┬──────┘
                                                          │
                                                          ▼
                                                   ┌─────────────┐
                                                   │   Logic     │
                                                   │ HandleFriend│
                                                   │   Request   │
                                                   └──────┬──────┘
                                                          │
                ┌─────────────────────────────────────────┼─────────────┐
                │                                         │             │
                ▼                                         ▼             ▼
         ┌─────────────┐                         ┌─────────────┐ ┌─────────────┐
         │ 查询申请    │                         │ 验证权限    │ │ 检查状态    │
         │   记录      │                         │ to_user_id  │ │ status=0?   │
         │             │                         │ == 当前用户 │ │             │
         └──────┬──────┘                         └──────┬──────┘ └──────┬──────┘
                │                                       │               │
                │                                       │               │
                ▼                                       ▼               ▼
         ┌─────────────┐                         ┌─────────────┐ ┌─────────────┐
         │ action=1?   │                         │ 是          │ │ 待处理      │
         │ (同意)      │                         │             │ │             │
         └──────┬──────┘                         └─────────────┘ └─────────────┘
                │ 是
                ▼
         ┌─────────────────────────────┐
         │  开启事务                    │
         │  1. 更新申请状态 = 1         │
         │  2. 插入 (A, B) 好友记录    │
         │  3. 插入 (B, A) 好友记录    │
         │  提交事务                    │
         └──────────┬──────────────────┘
                    │
                    ▼
         ┌─────────────────────┐
         │ 推送 WebSocket 通知  │
         │ 通知申请人(A)被同意  │
         └─────────────────────┘
```

**关键点**：
- ✅ **双向插入**：同意后插入两条好友记录
- 🔒 **事务保证**：确保数据一致性
- 📢 **实时通知**：通过 WebSocket 通知申请人

---

### 4.3 获取好友列表流程

```
┌─────────┐     GET /api/v1/friend/list?page=1    ┌─────────────┐
│  前端   │ ───────────────────────────────────► │  Handler    │
└─────────┘                                        └──────┬──────┘
                                                          │
                                                          ▼
                                                   ┌─────────────┐
                                                   │   Logic     │
                                                   │ GetFriend   │
                                                   │    List     │
                                                   └──────┬──────┘
                                                          │
                                                          ▼
                                                   ┌─────────────┐
                                                   │ 查询 Model  │
                                                   │  (带缓存)   │
                                                   └──────┬──────┘
                                                          │
                ┌─────────────────────────────────────────┼─────────────┐
                │                                         │             │
                ▼                                         ▼             ▼
         ┌─────────────┐                         ┌─────────────┐ ┌─────────────┐
         │ 查 Redis    │                         │ 未命中      │ │ 查 MySQL    │
         │ 缓存        │                         │             │ │             │
         └──────┬──────┘                         └─────────────┘ └──────┬──────┘
                │                                                       │
                │ 命中                                                  │
                ▼                                                       ▼
         ┌─────────────┐                                        ┌─────────────┐
         │ 返回缓存数据│                                        │ 写入缓存    │
         └─────────────┘                                        └──────┬──────┘
                                                                       │
                                                                       ▼
                                                                ┌─────────────┐
                                                                │ 返回结果    │
                                                                └─────────────┘
```

**性能优化**：
- 🚀 **Model 级缓存**：go-zero 自动实现
- 📄 **分页查询**：减少数据传输量
- 🔍 **索引优化**：`user_id` 作为查询条件

---

### 4.4 拉黑/取消拉黑流程

```
┌─────────┐     POST /api/v1/friend/blacklist     ┌─────────────┐
│  前端   │ ───────────────────────────────────► │  Handler    │
└─────────┘  {friendId, isBlack: true}            └──────┬──────┘
                                                          │
                                                          ▼
                                                   ┌─────────────┐
                                                   │   Logic     │
                                                   │ SetBlacklist│
                                                   └──────┬──────┘
                                                          │
                ┌─────────────────────────────────────────┴─────────────┐
                │                                                       │
                ▼                                                       ▼
         ┌─────────────┐                                        ┌─────────────┐
         │ 查询好友关系│                                        │ isBlack?    │
         │   是否存在  │                                        │             │
         └──────┬──────┘                                        └──────┬──────┘
                │                                                      │
                │ 存在                                                 │
                ▼                                                      ▼
         ┌─────────────────────────┐                         ┌──────────────────┐
         │ UPDATE im_friend        │                         │ true: status=2   │
         │ SET status = 2/1        │◄────────────────────────│ false: status=1  │
         │ WHERE user_id = ?       │                         └──────────────────┘
         │   AND friend_id = ?     │
         └──────────┬──────────────┘
                    │
                    ▼
         ┌─────────────────────┐
         │ 清除对应缓存         │
         └─────────────────────┘
```

**业务规则**：
- 🔗 **保留关系**：拉黑不删除好友记录，只改变 status
- 🚫 **单向控制**：只更新当前用户的记录
- 🔄 **可逆操作**：可以取消拉黑

---

## 五、核心组件详解

### 5.1 好友关系双向性

Friend 服务采用**双向记录**设计：

```go
// A 添加 B 为好友，需插入两条记录
record1: {user_id: A, friend_id: B, remark: "老王"}
record2: {user_id: B, friend_id: A, remark: ""}
```

**优点**：
- ✅ 查询效率高：直接通过 `user_id` 查询好友列表
- ✅ 备注独立：每个用户可设置不同备注
- ✅ 状态独立：拉黑只影响一方

**缺点**：
- ❌ 存储空间翻倍
- ❌ 需要事务保证一致性

---

### 5.2 Model 缓存机制

go-zero 自动生成的 Model 带 Redis 缓存：

```go
// 缓存 Key 设计
cacheImFriendIdPrefix = "cache:imFriend:id:"
cacheImFriendUserIdFriendIdPrefix = "cache:imFriend:userId:friendId:"

// 查询时自动缓存
func (m *defaultImFriendModel) FindOne(ctx context.Context, id int64) (*ImFriend, error) {
    // 1. 先查 Redis
    // 2. 未命中则查 MySQL
    // 3. 写入 Redis
    return friend, nil
}

// 更新时自动清除缓存
func (m *defaultImFriendModel) Update(ctx context.Context, data *ImFriend) error {
    // 1. 更新 MySQL
    // 2. 清除相关缓存
    return nil
}
```

---

### 5.3 Service Context（依赖注入）

```go
type ServiceContext struct {
    Config               config.Config
    ImFriendModel        model.ImFriendModel        // 好友 Model
    ImFriendRequestModel model.ImFriendRequestModel // 申请 Model
    FriendRpc            friendclient.Friend        // Friend RPC 客户端
}
```

---

## 六、API 层与 RPC 层关系

### 6.1 API 层职责

- 📥 **接收 HTTP 请求**
- 🔐 **JWT 认证**（go-zero 中间件）
- ✅ **参数验证**
- 🔄 **调用 RPC 层**
- 📤 **返回响应**

### 6.2 RPC 层职责

- 💼 **核心业务逻辑**
- 💾 **数据库操作**
- 🔔 **事件通知**（WebSocket 推送）
- 🔁 **事务管理**

### 6.3 调用流程示例

```go
// API Layer (Handler)
func (l *AddFriendRequestLogic) AddFriendRequest(req *types.AddFriendRequestReq) (*types.AddFriendRequestResp, error) {
    userId := l.ctx.Value("userId").(int64)
    
    // 调用 RPC
    resp, err := l.svcCtx.FriendRpc.AddFriendRequest(l.ctx, &friend.AddFriendRequestReq{
        UserId:   userId,
        FriendId: req.ToUserId,
        Message:  req.Message,
    })
    
    return &types.AddFriendRequestResp{RequestId: resp.RequestId}, nil
}

// RPC Layer (Logic)
func (l *AddFriendRequestLogic) AddFriendRequest(in *friend.AddFriendRequestReq) (*friend.AddFriendRequestResp, error) {
    // 1. 业务校验
    // 2. 数据库操作
    // 3. WebSocket 通知
    // 4. 返回结果
}
```

---

## 七、业务规则总结

### 好友申请规则

| 场景 | 是否允许 | 说明 |
|------|---------|------|
| 向自己发送申请 | ❌ | 不能添加自己为好友 |
| 已是好友 | ❌ | 已经是好友无需申请 |
| 已有待处理申请 | ❌ | 不能重复发送申请 |
| 对方拉黑我 | ✅ | 可以发送，但对方看不到 |

### 好友关系规则

| 操作 | 影响范围 | 说明 |
|------|---------|------|
| 删除好友 | 双向 | 双方都不再是好友 |
| 修改备注 | 单向 | 只影响自己的备注 |
| 拉黑好友 | 单向 | 只影响自己，对方仍可见我 |

---

## 八、性能优化策略

### 8.1 数据库优化

```sql
-- 索引优化
CREATE INDEX idx_to_user ON im_friend_request (to_user_id, status);
CREATE UNIQUE INDEX uk_user_friend ON im_friend (user_id, friend_id);

-- 分页查询优化
SELECT * FROM im_friend 
WHERE user_id = ? AND status = 1
LIMIT ?, ?;
```

### 8.2 缓存策略

- 🔥 **热点数据**：好友列表（缓存时间: 1小时）
- 🔄 **更新策略**：写时清除缓存（Cache Aside）
- 📦 **批量查询**：批量获取好友基本信息

### 8.3 接口优化

- 📄 **分页加载**：默认 20 条/页
- 🔍 **按需查询**：只返回必要字段
- 🚀 **异步通知**：WebSocket 推送不阻塞主流程

---

## 九、常见问题

### Q1: 为什么采用双向记录而不是单向？

**A**: 双向记录虽然占用更多存储空间，但查询效率更高。每个用户可以直接通过 `user_id` 查询好友列表，不需要 `OR` 查询或 `UNION`，且每个用户可以独立设置备注和拉黑状态。

### Q2: 删除好友后能否查看历史聊天记录？

**A**: 可以。好友关系和聊天记录是独立的，删除好友只删除 `im_friend` 表的记录，不影响 `im_message` 表。

### Q3: 拉黑和删除好友的区别？

**A**: 
- **删除好友**: 双方都不再是好友，需要重新申请
- **拉黑**: 好友关系保留，但对方无法给你发消息，你可以主动联系对方

### Q4: 如何防止恶意发送大量好友申请？

**A**: 可以在 Logic 层添加限流逻辑：
- 限制每天发送申请的数量
- 限制短时间内的申请频率
- 对被拒绝后的重复申请增加冷却时间

---

## 十、总结

Friend 微服务采用**分层架构**：

1. **Handler 层**：处理 HTTP 请求
2. **Logic 层**：核心业务逻辑
3. **Model 层**：数据访问（带缓存）
4. **RPC 层**：内部服务调用

**核心特点**：
- 📊 双向好友关系设计
- 🚀 自动缓存提升性能
- 🔒 事务保证数据一致性
- 📢 WebSocket 实时通知
- 🎯 职责清晰，易于维护

---

**文档作者**: Skylm  
**最后更新**: 2026-01-13  
**相关文档**: [Friend API 文档](./api/FRIEND_API文档.md)
