# Group 微服务架构与流程讲解

## 一、服务概述

Group 服务是 SkyeIM 即时通讯系统的**群组管理微服务**，负责群组的创建、成员管理、入群邀请、入群申请等核心社交功能。

### 技术栈

| 组件 | 技术选型 | 用途 |
|------|---------|------|
| 框架 | go-zero | 微服务框架 |
| 数据库 | MySQL | 群组关系持久化 |
| 缓存 | Redis | Model 缓存 |
| 通信 | gRPC + HTTP | RPC 内部调用 + HTTP 对外接口 |
| WebSocket | 推送通知 | 实时通知（邀请、申请）|

---

## 二、目录结构

```
app/group/
├── api/                                # HTTP AP层
│   ├── group.api                       # API 定义文件
│   └── internal/
│       ├── handler/                    # HTTP 处理器
│       │   ├── groupmgmt/              # 群组管理
│       │   ├── membermgmt/             # 成员管理
│       │   └── invitation/             # 邀请管理
│       └── logic/                      # 业务逻辑层
│           ├── groupmgmt/
│           ├── membermgmt/
│           └── invitation/
├── rpc/                                 # gRPC 服务层
│   └── internal/
│       └── logic/                      # RPC 业务逻辑
├── model/                               # 数据模型层
│   ├── im_group.sql                    # 群组表 DDL
│   ├── im_group_member.sql             # 群成员表 DDL
│   ├── im_group_invitation.sql         # 群邀请表 DDL
│   ├── im_group_join_request.sql       # 入群申请表 DDL
│   └── *.go                            # Model 实现
└── GROUP_API文档.md                     # 前端对接文档
```

---

## 三、数据库设计

### 3.1 群组表 (im_group)

```sql
CREATE TABLE `im_group` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `group_id` varchar(64) NOT NULL COMMENT '群组唯一标识',
  `name` varchar(100) NOT NULL COMMENT '群名称',
  `avatar` varchar(255) DEFAULT NULL COMMENT '群头像',
  `owner_id` bigint NOT NULL COMMENT '群主ID',
  `description` varchar(500) DEFAULT NULL COMMENT '群描述',
  `max_members` int DEFAULT 200 COMMENT '最大成员数',
  `member_count` int DEFAULT 0 COMMENT '当前成员数',
  `status` tinyint DEFAULT 1 COMMENT '状态: 1-正常 2-已解散',
  `invite_confirm_mode` tinyint DEFAULT 1 COMMENT '邀请确认模式: 0-直接加入 1-需要确认',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_id` (`group_id`),
  KEY `idx_owner_id` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**设计要点**：
- 🔑 **唯一索引** `uk_group_id`：群组ID全局唯一
- 👑 **owner_id**：群主用户 ID，一个群只有一个群主
- 📊 **member_count**：冗余字段，加速查询（避免 COUNT）
- 🚫 **status**：1-正常，2-已解散（软删除）
- ⚙️ **invite_confirm_mode**：控制邀请是否需要确认

### 3.2 群成员表 (im_group_member)

```sql
CREATE TABLE `im_group_member` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `group_id` varchar(64) NOT NULL COMMENT '群组ID',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `role` tinyint DEFAULT 3 COMMENT '角色: 1-群主 2-管理员 3-普通成员',
  `nickname` varchar(50) DEFAULT NULL COMMENT '群昵称',
  `mute` tinyint DEFAULT 0 COMMENT '是否禁言: 0-否 1-是',
  `read_seq` BIGINT UNSIGNED DEFAULT 0 COMMENT '已读消息Seq',
  `joined_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_user` (`group_id`, `user_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**设计要点**：
- 🔗 **唯一索引** `uk_group_user`：防止重复加入
- 👤 **role**：1-群主，2-管理员，3-普通成员
- 📖 **read_seq**：群聊已读进度（用于未读计数）
- 🔇 **mute**：禁言状态

### 3.3 群邀请表 (im_group_invitation)

```sql
CREATE TABLE `im_group_invitation` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `group_id` varchar(64) NOT NULL COMMENT '群组ID',
  `inviter_id` bigint unsigned NOT NULL COMMENT '邀请人ID',
  `invitee_id` bigint unsigned NOT NULL COMMENT '被邀请人ID',
  `message` varchar(200) DEFAULT '' COMMENT '邀请消息',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0-待处理 1-已同意 2-已拒绝 3-已过期',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_invitee_status` (`invitee_id`, `status`),
  KEY `idx_inviter` (`inviter_id`),
  KEY `idx_group_invitee` (`group_id`, `invitee_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**设计要点**：
- 📨 **单向邀请**：inviter → invitee
- 🔄 **状态流转**：0(待处理) → 1(已同意) / 2(已拒绝) / 3(已过期)
- 🔍 **索引优化**：`idx_invitee_status` 用于查询收到的待处理邀请

### 3.4 入群申请表 (im_group_join_request)

```sql
CREATE TABLE `im_group_join_request` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `group_id` varchar(64) NOT NULL COMMENT '群组ID',
  `user_id` bigint unsigned NOT NULL COMMENT '申请人ID',
  `message` varchar(200) DEFAULT '' COMMENT '申请理由',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '0-待处理 1-已同意 2-已拒绝',
  `handler_id` bigint unsigned DEFAULT NULL COMMENT '处理人ID（群主/管理员）',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_group_status` (`group_id`, `status`),
  KEY `idx_user` (`user_id`),
  UNIQUE KEY `uk_group_user_pending` (`group_id`, `user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**设计要点**：
- 🙋 **主动申请**：user → group
- 👮 **handler_id**：记录是哪个管理员/群主处理的
- 🔒 **唯一约束**：防止同一用户对同一群组有多个待处理申请

---

## 四、核心流程

### 4.1 创建群组流程

```
┌─────────┐     POST /api/v1/group/create      ┌─────────────┐
│  前端   │ ───────────────────────────────────► │  Handler    │
└─────────┘  {name, avatar, memberIds}          └──────┬──────┘
                                                       │
                                                       ▼
                                                ┌─────────────┐
                                                │   Logic     │
                                                │ CreateGroup │
                                                └──────┬──────┘
                                                       │
                ┌──────────────────────────────────────┼──────────────┐
                │                                      │              │
                ▼                                      ▼              ▼
         ┌─────────────┐                      ┌─────────────┐ ┌─────────────┐
         │ 生成群组ID  │                      │ 插入群组记录│ │ 插入成员记录│
         │ g_yyyyMMdd  │                      │ im_group    │ │ im_group_   │
         │ _序号       │                      │             │ │ member      │
         └──────┬──────┘                      └──────┬──────┘ └──────┬──────┘
                │                                    │              │
                │                                    │              │
                └────────────────┬───────────────────┘              │
                                 │                                  │
                                 ▼                                  ▼
                          ┌─────────────┐                  ┌─────────────┐
                          │ 创建者成为  │                  │ 初始成员    │
                          │ 群主(role=1)│                  │ role=3      │
                          └─────────────┘                  └─────────────┘
```

**关键逻辑**：
1. 生成全局唯一的 `group_id`（格式：`g_20260113_001`）
2. 插入 `im_group` 表
3. 插入创建者到 `im_group_member`，role=1（群主）
4. 如果有 `memberIds`，批量插入成员，role=3（普通成员）
5. 更新 `member_count`

---

### 4.2 入群邀请流程（需确认）

```
步骤1: 成员A邀请用户B
POST /api/v1/group/invitation/send
    ↓
插入 im_group_invitation
status=0（待处理）
    ↓
WebSocket推送给用户B
    ↓
用户B查看邀请
GET /api/v1/group/invitation/received
    ↓
用户B处理邀请
POST /api/v1/group/invitation/handle {action: 1}
    ↓
开启事务:
1. 更新邀请状态=1
2. 插入 im_group_member (user_id=B, role=3)
3. 更新 member_count +1
提交事务
    ↓
WebSocket推送给邀请人A
```

---

### 4.3 入群申请流程

```
步骤1: 用户搜索群组
GET /api/v1/group/search/precise?groupId=g_xxx
    ↓
显示群组信息，提供"申请加入"按钮
    ↓
步骤2: 用户发起申请
POST /api/v1/group/join/request
    ↓
检查:
1. 是否已是成员？
2. 是否有pending申请？
3. 群组是否存在且正常？
    ↓
插入 im_group_join_request
status=0（待处理）
    ↓
WebSocket推送给群主/管理员
    ↓
步骤3: 管理员查看申请
GET /api/v1/group/join/list
    ↓
管理员处理申请
POST /api/v1/group/join/handle {action: 1}
    ↓
开启事务:
1. 更新申请状态=1
2. 插入 im_group_member (role=3)
3. 更新 member_count +1
4. 记录 handler_id
提交事务
    ↓
WebSocket推送给申请人
```

---

### 4.4 解散群组流程

```
POST /api/v1/group/dismiss
    ↓
验证权限（仅群主）
    ↓
开启事务:
1. 更新 im_group status=2（已解散）
2. 删除所有 im_group_member
提交事务
    ↓
WebSocket推送给所有成员
    ↓
前端收到后移除群聊
```

---

## 五、四种入群方式对比

| 方式 | 接口 | 流程 | 场景 |
|------|------|------|------|
| **直接拉人** | POST /api/v1/group/member/invite | 直接加入，无需确认 | 好友间互相拉人 |
| **邀请确认** | POST /api/v1/group/invitation/send | 发送邀请 → 对方同意 → 加入 | 正式群组 |
| **主动申请** | POST /api/v1/group/join/request | 用户申请 → 管理员审批 → 加入 | 公开群组 |
| **创建时指定** | POST /api/v1/group/create | 创建群组时指定初始成员 | 讨论组 |

**配置项**：
```yaml
invite_confirm_mode:
  0: 直接加入（无需确认）
  1: 需要确认（默认）
```

---

## 六、核心组件详解

### 6.1 群组 ID 生成

**格式**：`g_yyyyMMdd_序号`

**示例**：
- `g_20260113_001`
- `g_20260113_002`

**实现**：
```go
func generateGroupId() string {
    date := time.Now().Format("20060102")
    // 查询当天最大序号
    seq := getMaxSeqToday() + 1
    return fmt.Sprintf("g_%s_%03d", date, seq)
}
```

**优点**：
- ✅ 可读性强（能看出创建日期）
- ✅ 全局唯一
- ✅ 支持分库分表（按日期）

---

### 6.2 成员角色权限

| 角色 | role值 | 权限 |
|------|-------|------|
| **群主** | 1 | 解散群组、转让群主、设置管理员、踢人、禁言、修改群信息 |
| **管理员** | 2 | 踢人（不含群主）、禁言、修改群信息、审批入群申请 |
| **普通成员** | 3 | 发言、邀请好友、退群 |

**权限检查示例**：
```go
func checkPermission(userId int64, groupId string, requiredRole int) error {
    member := getMember(userId, groupId)
    if member.Role > requiredRole {
        return errors.New("权限不足")
    }
    return nil
}

// 使用
err := checkPermission(userId, groupId, 2) // 需要管理员或群主权限
```

---

### 6.3 已读 Seq 机制

**概念**：
- 每条群消息有一个递增的 `Seq`
- 每个成员维护自己的 `read_seq`
- 未读数 = 最新Seq - read_seq

**流程**：
```
用户收到新消息 (seq=1250)
    ↓
用户阅读消息
    ↓
前端上报已读
POST /api/v1/group/read {readSeq: 1250}
    ↓
更新 im_group_member.read_seq = 1250
    ↓
下次查询未读时:
未读数 = 最新Seq(1300) - read_seq(1250) = 50条
```

---

### 6.4 WebSocket 实时通知

**通知场景**：

| 事件 | 推送对象 | 通知内容 |
|------|---------|---------|
| 收到邀请 | 被邀请人 | "XXX 邀请你加入 YYY 群组" |
| 邀请被同意 | 邀请人 | "XXX 已同意加入群组" |
| 收到申请 | 群主/管理员 | "XXX 申请加入群组" |
| 申请被同意 | 申请人 | "你的入群申请已通过" |
| 被踢出群 | 被踢成员 | "你已被移出群组" |
| 群组解散 | 所有成员 | "群组已解散" |

**实现**：
```go
// 发送邀请后
wsClient.PushNotification(inviteeId, Notification{
    Type: "group_invitation",
    Data: map[string]interface{}{
        "invitationId": invitationId,
        "groupId": groupId,
        "groupName": groupName,
        "inviterName": inviterName,
    },
})
```

---

## 七、业务规则总结

### 群组管理规则

| 操作 | 权限要求 | 特殊规则 |
|------|---------|---------|
| 创建群组 | 任何用户 | 创建者自动成为群主 |
| 解散群组 | 仅群主 | 删除所有成员记录 |
| 修改群信息 | 群主或管理员 | - |
| 退出群组 | 普通成员 | 群主不能退群（只能解散） |

### 成员管理规则

| 操作 | 权限要求 | 特殊规则 |
|------|---------|---------|
| 邀请成员 | 任何成员 | 根据 invite_confirm_mode 决定是否需要确认 |
| 踢出成员 | 群主或管理员 | 不能踢群主 |
| 设置管理员 | 仅群主 | - |
| 禁言成员 | 群主或管理员 | 不能禁言群主和管理员 |

### 入群规则

| 场景 | 重复申请 | 检查 |
|------|---------|------|
| 直接拉人 | 忽略已存在成员 | 检查是否超过 max_members |
| 邀请确认 | 允许重复邀请 | 检查邀请是否已存在且pending |
| 主动申请 | 不允许重复申请 | 唯一约束 `uk_group_user_pending` |

---

## 八、性能优化策略

### 8.1 数据库优化

```sql
-- 索引优化
CREATE INDEX idx_invitee_status ON im_group_invitation (invitee_id, status);
CREATE INDEX idx_group_status ON im_group_join_request (group_id, status);
CREATE UNIQUE INDEX uk_group_user ON im_group_member (group_id, user_id);

-- 分页查询优化
SELECT * FROM im_group_member 
WHERE group_id = ? 
LIMIT ?, ?;
```

### 8.2 冗余字段

- `member_count`：避免每次查询都 COUNT
- 更新策略：插入/删除成员时同步更新

### 8.3 缓存策略

- 🔥 **热点数据**：群信息（缓存时间: 1小时）
- 🔄 **更新策略**：写时清除缓存（Cache Aside）
- 📦 **批量查询**：批量获取群组基本信息

---

## 九、常见问题

### Q1: 为什么需要四种入群方式？

**A**: 不同场景有不同需求：
- **讨论组**：快速拉人，无需确认（直接拉人）
- **正式群组**：需要对方同意（邀请确认）
- **公开群组**：需要审核（主动申请）
- **初始创建**：批量添加（创建时指定）

### Q2: 群主如何转让？

**A**: 当前未实现。如需实现：
1. 添加 `TransferOwner` 接口
2. 更新 `im_group.owner_id`
3. 更新 `im_group_member`：旧群主 role=3，新群主 role=1

### Q3: 为什么入群申请表有唯一约束？

**A**: 防止用户多次申请刷屏。约束包含 `status` 字段，允许用户在上次申请被拒绝后重新申请。

### Q4: 如何实现群公告？

**A**: 需扩展 `im_group` 表：
```sql
ALTER TABLE im_group 
ADD COLUMN `announcement` varchar(1000) DEFAULT NULL COMMENT '群公告';
```

---

## 十、总结

Group 微服务采用**分层架构**：

1. **Handler 层**：处理 HTTP 请求
2. **Logic 层**：核心业务逻辑
3. **Model 层**：数据访问（带缓存）
4. **RPC 层**：内部服务调用

**核心特点**：
- 📊 四表设计（群组、成员、邀请、申请）
- 🚀 多种入群方式满足不同场景
- 🔒 完善的权限控制
- 📢 WebSocket 实时通知
- 📖 Seq 机制支持未读计数
- 🎯 职责清晰，易于维护

---

**文档作者**: Skylm  
**最后更新**: 2026-01-13  
**相关文档**: [Group API 文档](./GROUP_API文档.md)
