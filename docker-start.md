# Docker 部署指南

## 📦 前置要求

确保已安装：
- Docker Desktop (Windows/Mac) 或 Docker Engine (Linux)
- Docker Compose

检查版本：
```bash
docker --version
docker-compose --version
```

## 🚀 快速启动

### 1. 启动所有服务
```bash
# 在项目根目录执行
docker-compose up -d
```

### 2. 查看服务状态
```bash
docker-compose ps
```

### 3. 查看日志
```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f gateway
docker-compose logs -f ws-server
```

### 4. 停止服务
```bash
docker-compose down
```

### 5. 重新构建并启动
```bash
# 如果修改了代码，需要重新构建镜像
docker-compose up -d --build
```

## 🔧 配置说明

### 服务端口映射

| 服务 | 容器端口 | 主机端口 | 说明 |
|------|---------|---------|------|
| Gateway | 8080 | 8080 | API 网关 |
| Auth API | 10001 | 10001 | 认证服务 |
| User API | 10100 | 10100 | 用户服务 |
| Friend API | 10200 | 10200 | 好友服务 |
| Message API | 10400 | 10400 | 消息服务 |
| Group API | 10500 | 10500 | 群组服务 |
| Upload API | 10600 | 10600 | 上传服务 |
| WebSocket | 10300 | 10300 | WebSocket 服务 |
| MySQL | 3306 | 3306 | 数据库 |
| Redis | 6379 | 16379 | 缓存 |
| etcd | 2379 | 2379 | 服务发现 |
| MinIO | 9000 | 9000 | 对象存储 |
| MinIO Console | 9001 | 9001 | MinIO 管理界面 |

### 数据库配置

**MySQL:**
- 用户名: `root`
- 密码: `630630`
- 数据库: `im_auth`
- 端口: `3306`

**Redis:**
- 端口: `16379` (映射到容器的 6379)
- 无密码

**MinIO:**
- 用户名: `minioadmin`
- 密码: `minioadmin`
- API 端口: `9000`
- 控制台: `http://localhost:9001`

## ⚠️ 注意事项

### 1. 配置文件修改

Docker 部署时，需要修改各服务的配置文件，将 `127.0.0.1` 改为容器服务名：

**示例: `app/auth/etc/auth-api.yaml`**
```yaml
# 修改前
MySQL:
  DataSource: root:630630@tcp(127.0.0.1:3306)/im_auth

Cache:
  - Host: 127.0.0.1:16379

# 修改后
MySQL:
  DataSource: root:630630@tcp(mysql:3306)/im_auth

Cache:
  - Host: redis:6379
```

**需要修改的服务:**
- `app/auth/etc/auth-api.yaml`
- `app/user/api/etc/user-api.yaml`
- `app/friend/api/etc/friend-api.yaml`
- `app/message/api/etc/message-api.yaml`
- `app/group/api/etc/group-api.yaml`
- `app/upload/api/etc/upload-api.yaml`
- `app/ws/etc/ws.yaml`
- `app/gateway/etc/gateway.yaml`

**替换规则:**
- `127.0.0.1:3306` → `mysql:3306`
- `127.0.0.1:16379` → `redis:6379`
- `127.0.0.1:2379` → `etcd:2379`
- `127.0.0.1:9000` → `minio:9000`

### 2. 数据库初始化

首次启动时，需要手动导入数据库表结构：

```bash
# 等待 MySQL 启动完成（约 30 秒）
docker-compose logs mysql | grep "ready for connections"

# 进入 MySQL 容器
docker exec -it skyeim-mysql mysql -uroot -p630630 im_auth

# 或者从外部导入 SQL 文件
docker exec -i skyeim-mysql mysql -uroot -p630630 im_auth < your_schema.sql
```

### 3. 服务启动顺序

Docker Compose 已配置服务依赖关系，会自动按顺序启动：
1. 基础服务: etcd, Redis, MySQL, MinIO
2. 应用服务: Auth, User, Friend, Message, Group, Upload, WebSocket
3. 网关: Gateway (最后启动)

## 🐛 故障排查

### 服务无法启动
```bash
# 查看服务日志
docker-compose logs [service-name]

# 重启特定服务
docker-compose restart [service-name]
```

### 清理并重新开始
```bash
# 停止并删除所有容器
docker-compose down

# 删除数据卷（会清空数据库数据！）
docker-compose down -v

# 重新构建并启动
docker-compose up -d --build
```

### 查看容器内部
```bash
# 进入容器
docker exec -it skyeim-gateway sh

# 查看配置文件
docker exec skyeim-gateway cat /app/etc/gateway.yaml
```

## 📊 测试部署

启动成功后，测试接口：

```bash
# 测试网关
curl http://localhost:8080/health

# 测试认证服务
curl -X POST http://localhost:8080/api/v1/auth/captcha/send \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}'
```

## 🎯 生产环境建议

1. **修改默认密码**: MySQL、Redis、MinIO 的密码
2. **持久化数据**: 使用 Docker volumes 或外部存储
3. **资源限制**: 为每个服务设置 CPU 和内存限制
4. **日志管理**: 配置日志驱动，避免日志文件过大
5. **健康检查**: 已配置，但可以根据实际情况调整

## 📝 常用命令

```bash
# 启动
docker-compose up -d

# 停止
docker-compose down

# 重启
docker-compose restart

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f [service-name]

# 重新构建
docker-compose build [service-name]

# 扩展服务（例如启动 3 个 Gateway 实例）
docker-compose up -d --scale gateway=3
```
