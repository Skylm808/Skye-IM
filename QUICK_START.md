# 🚀 SkyeIM 快速开始指南

## 📦 5分钟快速部署

### 步骤 1：克隆项目

```bash
git clone https://github.com/Skylm808/SkyeIM.git
cd SkyeIM
```

### 步骤 2：启动服务（选择其一）

#### 🐳 方式一：Docker 一键启动（推荐）

**Windows:**
```bash
scripts\docker-deploy.bat start
```

**Linux/Mac:**
```bash
chmod +x scripts/docker-deploy.sh
./scripts/docker-deploy.sh start
```

#### 💻 方式二：本地开发启动

```bash
# 1. 启动基础服务（MySQL, Redis, etcd, MinIO）
# 请确保已安装并启动这些服务

# 2. 初始化数据库
mysql -u root -p < init_database.sql

# 3. 配置 QQ 邮箱 SMTP（见下方说明）

# 4. 启动所有服务
cd app/gateway && go run gateway.go &
cd app/auth && go run auth.go &
cd app/user/api && go run user.go &
cd app/user/rpc && go run user.go &
cd app/friend/api && go run friend.go &
cd app/friend/rpc && go run friend.go &
cd app/message/api && go run message.go &
cd app/message/rpc && go run message.go &
cd app/group/api && go run group.go &
cd app/group/rpc && go run group.go &
cd app/upload/api && go run upload.go &
cd app/ws && go run ws.go &
```

### 步骤 3：验证服务

#### 使用健康检查脚本

**Windows:**
```bash
scripts\health-check.bat
```

**Linux/Mac:**
```bash
./scripts/health-check.sh
```

#### 手动验证

```bash
# 检查 Gateway 是否运行
curl http://localhost:8080

# 应该返回类似：{"code":401,"msg":"Unauthorized"}
# 这说明 Gateway 正常运行
```

### 步骤 4：测试 API

#### 1. 发送验证码

```bash
curl -X POST http://localhost:8080/api/v1/auth/captcha/send \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your@email.com"
  }'
```

#### 2. 注册用户

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456",
    "email": "your@email.com",
    "captcha": "123456"
  }'
```

#### 3. 登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456"
  }'
```

响应会返回 `access_token` 和 `refresh_token`。

#### 4. 使用 Token 访问受保护接口

```bash
# 将 YOUR_ACCESS_TOKEN 替换为上一步获取的 token
curl http://localhost:8080/api/v1/user/info \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

---

## 📧 QQ 邮箱 SMTP 配置（本地部署需要）

> Docker 部署也需要配置邮箱才能发送验证码！

### 步骤 1：获取 QQ 邮箱授权码

1. 登录 [QQ邮箱](https://mail.qq.com)
2. 点击 **设置** → **账户**
3. 找到 **POP3/SMTP服务** 或 **IMAP/SMTP服务**
4. 点击 **开启**
5. 按照提示发送短信验证
6. 点击 **生成授权码**
7. **复制并保存授权码**（不是 QQ 密码！）

### 步骤 2：配置服务

#### Docker 部署修改：

编辑 `app/auth/etc/auth-api-docker.yaml`:

```yaml
Email:
  Host: smtp.qq.com
  Port: 465
  Username: your@qq.com          # 你的完整 QQ 邮箱
  Password: your-auth-code       # 刚才生成的授权码
  From: "SkyeIM系统"
```

重启服务：
```bash
docker-compose restart auth-api
```

#### 本地部署修改：

编辑 `app/auth/etc/auth-api.yaml`:

```yaml
Email:
  Host: smtp.qq.com
  Port: 465
  Username: your@qq.com
  Password: your-auth-code
  From: "SkyeIM系统"
```

重启 Auth 服务：
```bash
# 停止旧进程
pkill -f auth.go

# 重新启动
cd app/auth && go run auth.go
```

---

## 🔍 服务端口说明

| 服务 | 端口 | 说明 |
|------|------|------|
| Gateway | 8080 | API 网关（统一入口） |
| Auth API | 10001 | 认证服务 |
| User API | 10100 | 用户管理 API |
| User RPC | 9100 | 用户服务 RPC |
| Friend API | 10200 | 好友管理 API |
| Friend RPC | 9200 | 好友服务 RPC |
| WebSocket | 10300 | 实时通信 |
| Message API | 10400 | 消息管理 API |
| Message RPC | 9300 | 消息服务 RPC |
| Group API | 10500 | 群组管理 API |
| Group RPC | 9400 | 群组服务 RPC |
| Upload API | 10600 | 文件上传 |
| MySQL | 3306 | 数据库 |
| Redis | 16379 | 缓存 |
| etcd | 2379 | 服务发现 |
| MinIO | 9000 | 对象存储 |
| MinIO Console | 9001 | MinIO 管理界面 |

---

## 🛠 常用命令

### Docker 部署

```bash
# 启动所有服务
scripts\docker-deploy.bat start        # Windows
./scripts/docker-deploy.sh start       # Linux/Mac

# 查看服务状态
scripts\docker-deploy.bat status
./scripts/docker-deploy.sh status

# 查看日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f gateway
docker-compose logs -f user-rpc

# 重启服务
docker-compose restart

# 重启特定服务
docker-compose restart gateway

# 停止所有服务
scripts\docker-deploy.bat stop
./scripts/docker-deploy.sh stop

# 停止并删除数据（⚠️ 危险操作）
scripts\docker-deploy.bat clean
./scripts/docker-deploy.sh clean
```

### 本地部署

```bash
# 查看进程
ps aux | grep go

# 停止所有 Go 服务
pkill -f "go run"

# 查看端口占用
netstat -ano | findstr :8080     # Windows
lsof -i :8080                    # Linux/Mac

# 杀死占用端口的进程
# Windows: taskkill /PID <PID> /F
# Linux/Mac: kill -9 <PID>
```

---

## 📱 前端项目

后端启动成功后，可以配合前端项目使用：

**前端仓库**: [Skye-IM-Front](https://github.com/Skylm808/Skye-IM-Front)

```bash
# 克隆前端项目
git clone https://github.com/Skylm808/Skye-IM-Front.git
cd Skye-IM-Front

# 安装依赖
npm install

# 启动开发服务器
npm start
```

前端默认会连接到 `http://localhost:8080` 的后端服务。

---

## ❓ 常见问题

### 1. 端口已被占用

**错误信息**: `bind: address already in use`

**解决方法**:
```bash
# 查看占用端口的进程
netstat -ano | findstr :8080     # Windows
lsof -i :8080                    # Linux/Mac

# 杀死进程或修改配置文件中的端口
```

### 2. Docker 服务启动失败

**解决方法**:
```bash
# 查看服务日志
docker-compose logs service-name

# 重新构建并启动
docker-compose up -d --build --force-recreate
```

### 3. 邮件发送失败

**常见原因**:
- ❌ 使用了 QQ 密码而不是授权码
- ❌ 端口配置错误（必须是 465）
- ❌ 未开启 SMTP 服务

**解决方法**:
- 检查 `auth-api.yaml` 或 `auth-api-docker.yaml` 配置
- 确保使用授权码（不是密码）
- 端口必须是 `465`

### 4. 数据库连接失败

**Docker 部署**:
```bash
# 等待 MySQL 容器完全启动
docker-compose logs mysql

# 进入 MySQL 容器测试
docker-compose exec mysql mysql -uroot -p630630
```

**本地部署**:
```bash
# 测试 MySQL 连接
mysql -h 127.0.0.1 -P 3306 -u root -p

# 检查数据库是否存在
SHOW DATABASES;
USE im_auth;
SHOW TABLES;
```

### 5. etcd 连接失败

```bash
# Docker 部署
docker-compose logs etcd

# 本地部署 - 测试 etcd
curl http://127.0.0.1:2379/version
```

---

## 📚 下一步

- 📖 查看 [API 文档](./API/) 了解接口详情
- 🏗️ 阅读 [架构文档](./docs/) 理解系统设计
- 🐳 查看 [Docker 部署文档](./DOCKER_DEPLOYMENT.md) 了解更多配置
- 💬 体验完整的 IM 功能（消息、好友、群组）

---

## 🆘 获取帮助

如果遇到问题：

1. 查看 [README.md](./README.md) 主文档
2. 查看 [DOCKER_DEPLOYMENT.md](./DOCKER_DEPLOYMENT.md) Docker 详细文档
3. 在 GitHub 提交 [Issue](https://github.com/Skylm808/SkyeIM/issues)

---

**祝你使用愉快！** 🎉

