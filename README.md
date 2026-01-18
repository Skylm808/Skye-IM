# SkyeIM - 现代化即时通讯系统

<div align="center">

![Go Version](https://img.shields.io/badge/Go-1.25.4-blue)
![Framework](https://img.shields.io/badge/Framework-go--zero-brightgreen)
![License](https://img.shields.io/badge/License-MIT-yellow)
![Architecture](https://img.shields.io/badge/Architecture-Microservice-orange)

一个基于 go-zero 框架构建的现代化即时通讯系统，采用微服务架构设计，支持私聊、群聊、好友管理等核心功能。

[功能特性](#功能特性) • [技术栈](#技术栈) • [快速开始](#快速开始) • [架构设计](#架构设计) • [API 文档](#api-文档) • [开发指南](#开发指南)

</div>

> [!NOTE]
> **前端项目**：[Skye-IM-Front](https://github.com/Skylm808/Skye-IM-Front) - 基于 React + Ant Design 的现代化 IM 客户端

---

## ✨ 功能特性

### 🔐 用户认证
- ✅ 邮箱验证码注册/登录
- ✅ JWT 双 Token 机制 (AccessToken + RefreshToken)
- ✅ 密码 bcrypt 加密存储
- ✅ 多方式登录（用户名/邮箱/手机号）
- ✅ Token 自动刷新机制

### 👥 好友管理
- ✅ 好友申请与处理
- ✅ 好友列表查询（分页）
- ✅ 好友删除
- ✅ 黑名单管理

### 💬 即时消息
- ✅ WebSocket 实时通信
- ✅ 私聊消息收发
- ✅ 群聊消息收发
- ✅ @提及功能
- ✅ 消息已读/未读状态
- ✅ 离线消息推送（WebSocket 连接时自动推送）
- ✅ 历史消息分页拉取（HTTP API）
- ✅ 模糊搜索聊天记录

### 👬 群组功能
- ✅ 创建/解散群组
- ✅ 群成员管理（邀请/踢出）
- ✅ 入群申请/审批
- ✅ 退出群聊
- ✅ 群信息修改
- ✅ 群组搜索

### 📁 文件管理
- ✅ 头像上传
- ✅ 文件上传下载
- ✅ MinIO 对象存储集成

### 🔍 搜索功能
- ✅ 精确搜索用户（用户名/邮箱/手机）
- ✅ 模糊搜索群组
- ✅ 消息内容搜索

### 👤 用户信息
- ✅ 个人资料管理
- ✅ 在线状态管理
- ✅ 个性签名/性别/地区设置

---

## 🛠 技术栈

### 后端框架
| 技术 | 版本 | 说明 |
|------|------|------|
| **语言** | Go 1.25.4 | 高性能编程语言 |
| **框架** | [go-zero](https://github.com/zeromicro/go-zero) 1.6.0 | 微服务框架 |
| **通信** | gRPC / HTTP / WebSocket | 多协议支持 |

### 存储层
| 技术 | 说明 |
|------|------|
| **数据库** | MySQL | 关系型数据存储 |
| **缓存** | Redis | 高速缓存、验证码存储 |
| **服务发现** | etcd | 分布式配置与服务注册 |
| **对象存储** | MinIO | 文件存储服务 |

### 核心依赖
```go
github.com/zeromicro/go-zero     // 微服务框架
github.com/golang-jwt/jwt/v4     // JWT 认证
golang.org/x/crypto              // 密码加密 (bcrypt)
gopkg.in/gomail.v2               // 邮件发送
github.com/minio/minio-go/v7     // MinIO SDK
```

---

## 🚀 快速开始

### 前置要求

- Go 1.25.4+
- MySQL 8.0+
- Redis 6.0+ (默认端口 16379)
- etcd 3.5+
- MinIO (可选，用于文件存储，默认端口 9000)

### 环境准备

#### 1️⃣ 安装依赖

```bash
# 克隆项目
git clone https://github.com/Skylm808/SkyeIM.git
cd SkyeIM

# 下载依赖
go mod download
```

#### 2️⃣ 启动基础服务

```bash
# 启动 MySQL (端口 3306)
# 创建数据库: im_auth (统一数据库，包含所有表)

# 启动 Redis (端口 16379)
redis-server --port 16379

# 启动 etcd (端口 2379)
etcd

# 启动 MinIO (端口 9000，可选)
minio server /data --console-address ":9001"
```

#### 3️⃣ 配置服务

修改各服务的配置文件 `etc/*.yaml`，配置数据库、Redis、etcd 连接信息。

**关键配置项**：
- **MySQL 连接字符串**: `root:630630@tcp(127.0.0.1:3306)/im_auth`
- **Redis 地址**: `127.0.0.1:16379` (无密码)
- **etcd 地址**: `127.0.0.1:2379`
- **JWT Secret**: `Skylm-im-secret-key` (所有服务必须保持一致)
- **数据库**: 统一使用 `im_auth` 数据库，包含以下表：
  - `user` - 用户信息
  - `im_friend` - 好友关系
  - `im_friend_request` - 好友申请
  - `im_message` - 消息记录
  - `im_group` - 群组信息
  - `im_group_member` - 群成员
  - `im_group_invitation` - 群邀请
  - `im_group_join_request` - 入群申请

### 启动服务

#### 方式一：独立启动各服务

```bash
# 1. 启动 Auth API (端口 10001)
cd app/auth && go run auth.go

# 2. 启动 User API (端口 10100)
cd app/user/api && go run user.go

# 3. 启动 Friend API (端口 10200)
cd app/friend/api && go run friend.go

# 4. 启动 Message API (端口 10400)
cd app/message/api && go run message.go

# 5. 启动 Group API (端口 10500)
cd app/group/api && go run group.go

# 6. 启动 Upload API (端口 10600)
cd app/upload/api && go run upload.go

# 7. 启动 WebSocket 服务 (端口 10300)
cd app/ws && go run ws.go

# 8. 启动 API 网关 (端口 8080)
cd app/gateway && go run gateway.go
```

#### 方式二：启动 RPC 服务（可选）

```bash
# User RPC (端口 9100)
cd app/user/rpc && go run user.go

# Friend RPC (端口 9200)
cd app/friend/rpc && go run friend.go

# Message RPC (端口 9300)
cd app/message/rpc && go run message.go

# Group RPC (端口 9400)
cd app/group/rpc && go run group.go
```

### 验证服务

```bash
# 测试注册接口
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456",
    "email": "test@example.com",
    "captcha": "123456"
  }'

# 测试登录接口
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456"
  }'
```

---

## 📝 配置示例

以下是核心服务的配置文件示例，帮助你快速配置项目。

<details>
<summary><b>点击查看 Auth API 配置</b> (app/auth/etc/auth-api.yaml)</summary>

```yaml
Name: auth-api
Port: 10001
MySQL:
  DataSource: root:630630@tcp(127.0.0.1:3306)/im_auth?charset=utf8mb4&parseTime=True
Cache:
  - Host: 127.0.0.1:16379
    Type: node
    Pass: ""
Auth:
  AccessSecret: "Skylm-im-secret-key"
  AccessExpire: 604800
Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: auth-api
```

</details>

<details>
<summary><b>点击查看 Gateway 配置</b> (app/gateway/etc/gateway.yaml)</summary>

```yaml
Name: gateway
Port: 8080
Etcd:
  Hosts:
    - 127.0.0.1:2379
Auth:
  AccessSecret: "Skylm-im-secret-key"
WhiteList:
  - ^/api/v1/auth/login$
  - ^/api/v1/auth/register$
  - ^/api/v1/auth/captcha/send$
```

</details>

<details>
<summary><b>点击查看 WebSocket 配置</b> (app/ws/etc/ws.yaml)</summary>

```yaml
Name: ws-server
Port: 10300
Auth:
  AccessSecret: "Skylm-im-secret-key"
Redis:
  Host: 127.0.0.1:16379
  Pass: ""
WebSocket:
  PingInterval: 30
  MaxMessageSize: 65536
```

</details>

> [!IMPORTANT]
> **配置要点**：
> - 🔑 JWT Secret: `Skylm-im-secret-key` (所有服务必须一致)
> - 🗄️ Redis 端口: `16379`
> - 🔐 MySQL: `root:630630@tcp(127.0.0.1:3306)/im_auth`
> - 📡 etcd: `127.0.0.1:2379`

---

## 🏗 架构设计

### 整体架构

```
┌─────────────┐
│   前端应用   │
└──────┬──────┘
       │ HTTP/WebSocket
       ▼
┌─────────────────────────────────────┐
│       API Gateway (8080)            │
│  ┌──────────┬───────────────────┐  │
│  │ JWT 鉴权 │ 服务发现 (etcd)   │  │
│  └──────────┴───────────────────┘  │
└─────────────────┬───────────────────┘
                  │
      ┌───────────┼───────────┬───────────┐
      ▼           ▼           ▼           ▼
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
│ Auth API │ │ User API │ │Friend API│ │Group API │
│  :10001  │ │  :10100  │ │  :10200  │ │  :10500  │
└────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘
     │            │            │            │
     └────────────┴────────────┴────────────┘
                  │
          ┌───────┴────────┐
          ▼                ▼
    ┌──────────┐     ┌──────────┐
    │  MySQL   │     │  Redis   │
    └──────────┘     └──────────┘

    ┌──────────────────────────┐
    │   WebSocket Server       │
    │        :10300            │
    └──────────────────────────┘
```

### 服务清单

| 服务 | 类型 | 端口 | 说明 |
|------|------|------|------|
| **gateway** | HTTP | 8080 | API 网关，统一入口 |
| **auth** | API | 10001 | 用户认证服务 |
| **user** | API/RPC | 10100/9100 | 用户信息服务 |
| **friend** | API/RPC | 10200/9200 | 好友管理服务 |
| **message** | API/RPC | 10400/9300 | 消息服务 |
| **group** | API/RPC | 10500/9400 | 群组服务 |
| **upload** | API | 10600 | 文件上传服务 |
| **ws** | WebSocket | 10300 | WebSocket 实时通信 |

### 技术特点

#### 🎯 微服务架构
- **服务隔离**：每个功能模块独立部署，互不影响
- **弹性扩展**：可根据负载独立扩展各服务
- **技术异构**：不同服务可选择最适合的技术栈

#### 🔄 服务通信
- **API 层**：HTTP RESTful API，面向前端
- **RPC 层**：gRPC 高性能内部调用
- **WebSocket**：实时双向通信

#### 🛡 网关层
- **统一入口**：所有请求经过 Gateway (8080)
- **JWT 鉴权**：双重 Token 验证机制
- **服务发现**：基于 etcd 自动路由
- **CORS 支持**：跨域配置

#### 📡 消息推送机制
- **WebSocket 连接**：
  - 连接成功后自动推送离线消息（私聊 + 群聊）
  - 实时接收新消息
- **HTTP API**：
  - 私聊历史：`GET /api/v1/message/history`
  - 群聊历史：`GET /api/v1/message/group/history`
  - 私聊离线同步：`GET /api/v1/message/offline`
  - 群聊离线同步：`GET /api/v1/message/group/sync`

#### 💾 数据层
- **MySQL**：持久化存储，支持事务
- **Redis 缓存**：
  - Model 级缓存（go-zero 自动）
  - 验证码存储（5 分钟 TTL）
  - 会话管理

---

## 📖 API 文档

### 认证服务 (Auth)

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 发送验证码 | POST | `/api/v1/auth/captcha/send` | 发送邮箱验证码 |
| 注册 | POST | `/api/v1/auth/register` | 用户注册 |
| 登录 | POST | `/api/v1/auth/login` | 用户登录 |
| 刷新 Token | POST | `/api/v1/auth/refresh` | 刷新访问令牌 |
| 退出登录 | POST | `/api/v1/auth/logout` | 用户登出 |
| 获取用户信息 | GET | `/api/v1/auth/userinfo` | 获取当前用户信息 |

### 好友服务 (Friend)

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 发送好友申请 | POST | `/api/v1/friend/request` | 发送好友申请 |
| 处理好友申请 | POST | `/api/v1/friend/request/handle` | 接受/拒绝好友申请 |
| 好友列表 | GET | `/api/v1/friend/list` | 获取好友列表 |
| 删除好友 | DELETE | `/api/v1/friend/delete` | 删除好友 |
| 设置备注 | PUT | `/api/v1/friend/remark` | 设置好友备注 |
| 黑名单管理 | POST/GET | `/api/v1/friend/blacklist` | 拉黑/查看黑名单 |

### 群组服务 (Group)

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 创建群组 | POST | `/api/v1/group/create` | 创建新群组 |
| 解散群组 | DELETE | `/api/v1/group/dismiss` | 解散群组（群主） |
| 退出群组 | POST | `/api/v1/group/quit` | 退出群聊 |
| 邀请成员 | POST | `/api/v1/group/invite` | 邀请用户入群 |
| 踢出成员 | DELETE | `/api/v1/group/kick` | 踢出群成员 |
| 入群申请 | POST | `/api/v1/group/join/request` | 申请加入群聊 |
| 处理申请 | POST | `/api/v1/group/join/handle` | 处理入群申请 |
| 群组搜索 | GET | `/api/v1/group/search` | 搜索群组 |

### 消息服务 (Message)

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 发送消息 | POST | `/api/v1/message/send` | 发送私聊消息 |
| 发送群消息 | POST | `/api/v1/message/group/send` | 发送群聊消息 |
| 消息列表 | GET | `/api/v1/message/list` | 获取消息历史 |
| 标记已读 | POST | `/api/v1/message/read` | 标记消息已读 |
| 未读消息 | GET | `/api/v1/message/unread` | 获取未读消息 |
| 搜索消息 | GET | `/api/v1/message/search` | 模糊搜索消息 |
| @我的消息 | GET | `/api/v1/message/at-me` | 获取@我的消息 |

### WebSocket 连接

```javascript
// 连接 WebSocket (端口 10300)
const ws = new WebSocket('ws://localhost:10300/ws?token=YOUR_ACCESS_TOKEN');

// 监听消息
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('收到消息:', data);
};

// 发送消息
ws.send(JSON.stringify({
  type: 'chat',
  to_user_id: 123,
  content: 'Hello!'
}));
```

---

## 👨‍💻 开发指南

### 项目结构

```
SkyeIM/
├── app/                    # 应用服务
│   ├── auth/              # 认证服务
│   ├── user/              # 用户服务
│   │   ├── api/          # HTTP API
│   │   ├── rpc/          # gRPC 服务
│   │   └── model/        # 数据模型
│   ├── friend/            # 好友服务
│   ├── message/           # 消息服务
│   ├── group/             # 群组服务
│   ├── upload/            # 上传服务
│   ├── ws/                # WebSocket 服务
│   └── gateway/           # API 网关
├── common/                # 公共组件
│   ├── captcha/          # 验证码
│   ├── email/            # 邮件发送
│   ├── jwt/              # JWT 工具
│   ├── errorx/           # 错误处理
│   └── response/         # 响应封装
├── docs/                  # 文档
├── go.mod                # Go 模块
└── README.md             # 项目说明
```

### 添加新服务

#### 1. 定义 API

创建 `.api` 文件定义接口：

```go
// example.api
syntax = "v1"

type ExampleReq {
    Name string `json:"name"`
}

type ExampleResp {
    Id   int64  `json:"id"`
    Name string `json:"name"`
}

@server(
    prefix: /api/v1/example
    jwt: Auth
)
service example-api {
    @handler Example
    post /create (ExampleReq) returns (ExampleResp)
}
```

#### 2. 生成代码

```bash
# 生成 API 代码
goctl api go -api example.api -dir .

# 生成 RPC 代码 (如果需要)
goctl rpc protoc example.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

#### 3. 实现业务逻辑

在 `internal/logic/` 目录实现业务逻辑。

#### 4. 配置服务发现

在 `etc/` 配置文件中添加 etcd 配置：

```yaml
Name: example-api
Port: 10700

Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: example-api
```

### 代码规范

- **命名**：遵循 Go 官方命名规范
- **分层**：严格遵循 Handler → Logic → Model
- **错误处理**：使用 `common/errorx` 统一错误码
- **日志**：使用 `logx` 记录关键操作
- **缓存**：合理使用 Redis 缓存

### 测试

```bash
# 单元测试
go test ./...

# 性能测试
go test -bench=. -benchmem
```

---

## 📋 TODO

- [ ] 实现消息撤回功能
- [ ] 添加语音/视频通话
- [ ] 实现端到端加密
- [ ] 添加消息已读回执
- [ ] 实现文件断点续传
- [ ] 添加 Prometheus 监控
- [ ] 实现分布式链路追踪
- [ ] Docker 容器化部署
- [ ] Kubernetes 编排

---

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

---

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

---

## 👤 作者

**Skylm**

- GitHub: [@Skylm808](https://github.com/Skylm808)

---

## 🙏 致谢

- [go-zero](https://github.com/zeromicro/go-zero) - 优秀的微服务框架
- [etcd](https://github.com/etcd-io/etcd) - 可靠的分布式键值存储
- [MinIO](https://github.com/minio/minio) - 高性能对象存储

---

<div align="center">

**如果这个项目对你有帮助，请给一个 ⭐️ Star 支持一下！**

Made with ❤️ by Skylm

</div>
