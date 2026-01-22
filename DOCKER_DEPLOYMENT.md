# SkyeIM Docker 完整部署指南

## 📋 架构概览

本项目采用完整的容器化微服务架构，所有服务都运行在 Docker 容器中。

### 服务列表

#### 基础设施服务
- **etcd** (2379, 2380) - 服务发现与配置中心
- **redis** (16379) - 缓存和会话存储
- **mysql** (3306) - 关系型数据库
- **minio** (9000, 9001) - 对象存储服务

#### RPC 服务（内部微服务）
- **user-rpc** (9100) - 用户RPC服务
- **friend-rpc** (9200) - 好友关系RPC服务
- **message-rpc** (9300) - 消息RPC服务
- **group-rpc** (9400) - 群组RPC服务

#### API 服务（HTTP接口）
- **auth-api** (10001) - 认证服务
- **user-api** (10100) - 用户管理服务
- **friend-api** (10200) - 好友管理服务
- **message-api** (10400) - 消息管理服务
- **group-api** (10500) - 群组管理服务
- **upload-api** (10600) - 文件上传服务

#### 其他服务
- **ws-server** (10300) - WebSocket 长连接服务
- **gateway** (8080) - API 网关（统一入口）

## 🚀 快速开始

### 1. 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少 4GB 可用内存
- 至少 10GB 可用磁盘空间

### 2. 构建并启动所有服务

```bash
# 构建并启动所有服务
docker-compose up -d --build

# 查看服务状态
docker-compose ps

# 查看服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f gateway
docker-compose logs -f user-rpc
```

### 3. 验证服务状态

```bash
# 检查所有容器是否正常运行
docker-compose ps

# 应该看到所有服务状态为 Up
```

### 4. 访问服务

- **API网关**: http://localhost:8080
- **MinIO 控制台**: http://localhost:9001
  - 用户名: minioadmin
  - 密码: minioadmin

## 📦 服务依赖关系

```
基础设施层:
├── etcd (服务发现)
├── redis (缓存)
├── mysql (数据库)
└── minio (对象存储)

RPC服务层:
├── user-rpc      → 依赖: mysql, redis, etcd
├── friend-rpc    → 依赖: mysql, redis, etcd
├── group-rpc     → 依赖: mysql, redis, etcd
└── message-rpc   → 依赖: mysql, redis, etcd, group-rpc

应用服务层:
├── auth-api      → 依赖: mysql, redis
├── user-api      → 依赖: mysql, redis, etcd, user-rpc
├── friend-api    → 依赖: mysql, redis, etcd, friend-rpc, user-rpc
├── message-api   → 依赖: mysql, redis, etcd, message-rpc, user-rpc, friend-rpc
├── group-api     → 依赖: mysql, redis, etcd, group-rpc, user-rpc
├── upload-api    → 依赖: mysql, redis, minio
└── ws-server     → 依赖: redis, etcd, user-rpc, friend-rpc, message-rpc, group-rpc

网关层:
└── gateway       → 依赖: 所有API服务 + 所有RPC服务
```

## 🔧 常用命令

### 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 启动特定服务
docker-compose up -d gateway user-rpc

# 重新构建并启动
docker-compose up -d --build
```

### 停止服务

```bash
# 停止所有服务
docker-compose down

# 停止并删除数据卷（⚠️ 会删除所有数据）
docker-compose down -v

# 停止特定服务
docker-compose stop gateway
```

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f user-rpc
docker-compose logs -f gateway

# 查看最近100行日志
docker-compose logs --tail=100 gateway
```

### 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart gateway
docker-compose restart user-rpc
```

### 查看服务状态

```bash
# 查看所有服务状态
docker-compose ps

# 查看服务资源使用
docker stats
```

## 🐛 故障排查

### 1. 服务启动失败

```bash
# 查看服务日志
docker-compose logs service-name

# 常见问题：
# - 端口被占用：修改 docker-compose.yaml 中的端口映射
# - 依赖服务未就绪：等待健康检查通过
# - 配置错误：检查 etc/*-docker.yaml 配置文件
```

### 2. 数据库连接失败

```bash
# 检查 MySQL 是否就绪
docker-compose logs mysql

# 进入 MySQL 容器
docker-compose exec mysql mysql -uroot -p630630

# 验证数据库是否创建
SHOW DATABASES;
```

### 3. 服务发现失败

```bash
# 检查 etcd 状态
docker-compose logs etcd

# 查看 etcd 中的服务注册信息
docker-compose exec etcd etcdctl get --prefix ""
```

### 4. Redis 连接问题

```bash
# 测试 Redis 连接
docker-compose exec redis redis-cli ping

# 查看 Redis 信息
docker-compose exec redis redis-cli info
```

## 📝 配置说明

### 环境变量配置

每个服务都有两个配置文件：
- `etc/service-name.yaml` - 本地开发配置
- `etc/service-name-docker.yaml` - Docker 容器配置

Docker 配置文件使用容器网络的服务名而不是 localhost：
- MySQL: `mysql:3306`
- Redis: `redis:6379`
- etcd: `etcd:2379`

### 修改配置

如需修改配置：
1. 编辑对应的 `*-docker.yaml` 文件
2. 重新构建并启动服务：
   ```bash
   docker-compose up -d --build service-name
   ```

## 🔐 安全建议

生产环境部署时，请务必：

1. **修改默认密码**
   - MySQL root 密码
   - Redis 密码（添加认证）
   - MinIO 管理员密码

2. **启用 HTTPS**
   - 在网关层配置 SSL 证书
   - 使用 Nginx 作为反向代理

3. **网络隔离**
   - 只暴露必要的端口（通常只需要暴露 gateway 的 8080）
   - 内部服务使用 Docker 内部网络通信

4. **添加限流和防护**
   - API 限流
   - DDoS 防护
   - 请求验证

## 📊 监控和日志

### 日志管理

建议配置集中式日志收集：
```bash
# 使用 ELK 或 Loki 收集容器日志
docker-compose logs -f | tee app.log
```

### 健康检查

所有服务都配置了健康检查：
- MySQL: `mysqladmin ping`
- Redis: `redis-cli ping`
- etcd: `etcdctl endpoint health`
- MinIO: `curl /minio/health/live`

## 🎯 开发建议

### 本地开发

如果你想在本地开发某个服务，可以：
1. 停止该服务的容器
2. 在本地启动该服务（使用 *-docker.yaml 配置）
3. 确保可以连接到 Docker 网络中的其他服务

```bash
# 停止某个服务
docker-compose stop user-api

# 在本地运行（需要能访问 Docker 网络）
cd app/user/api
go run user.go -f etc/user-api.yaml
```

### 热重载

对于频繁修改的服务，可以使用 volume 挂载源代码并配置热重载。

## 📚 更多信息

- [API 文档](./API/)
- [架构文档](./docs/)
- [面试文档](./interview_docs/)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

[LICENSE](./LICENSE)

