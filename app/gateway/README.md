# Gateway 网关服务

## 简介

SkyeIM的统一API网关，提供：
- ✅ 统一入口（8080端口）
- ✅ JWT鉴权（双重验证）
- ✅ etcd服务发现
- ✅ 反向代理转发
- ✅ CORS跨域支持

## 启动前准备

### 1. 启动etcd
```bash
etcd
```

### 2. 启动后端API服务（需要注册到etcd）
```bash
# Auth API
cd app/auth && go run auth.go

# User API  
cd app/user/api && go run user-api.go

# Friend API
cd app/friend/api && go run friend-api.go

# Message API
cd app/message/api && go run message-api.go
```

**重要**：确保所有API服务已配置etcd并成功注册：
- auth-api → 127.0.0.1:10000
- user-api → 127.0.0.1:10100
- friend-api → 127.0.0.1:10200
- message-api → 127.0.0.1:10400

## 启动Gateway

```bash
cd app/gateway
go run gateway.go
```

启动成功输出：
```
Gateway 启动在端口: 8080
支持的服务: auth-api, user-api, friend-api, message-api
已连接etcd: [127.0.0.1:2379]
```

## 测试接口

### 1. 公开接口（无需Token）

```bash
# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'

# 注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"newuser","password":"123456","email":"test@test.com","captcha":"123456"}'
```

### 2. 需要认证的接口（需要Token）

```bash
# 获取用户资料
curl -X GET http://localhost:8080/api/v1/user/profile \
  -H "Authorization: Bearer {your_token}"

# 获取好友列表
curl -X GET "http://localhost:8080/api/v1/friend/list?page=1&pageSize=20" \
  -H "Authorization: Bearer {your_token}"

# 获取消息历史
curl -X GET "http://localhost:8080/api/v1/message/history?peerId=123" \
  -H "Authorization: Bearer {your_token}"
```

## 路由规则

Gateway根据URL自动路由到对应服务：

| URL Pattern | 路由到 | 端口 |
|------------|--------|------|
| `/api/v1/auth/*` | auth-api | 10000 |
| `/api/v1/user/*` | user-api | 10100 |
| `/api/v1/friend/*` | friend-api | 10200 |
| `/api/v1/message/*` | message-api | 10400 |

## 鉴权机制

**双重验证（推荐配置）**：

1. **Gateway层**：验证JWT，成功后注入用户信息到Header
   - `X-User-Id`: 用户ID
   - `X-Username`: 用户名

2. **后端API层**：继续使用JWT中间件验证（保留现有代码）

**白名单**（不需要鉴权的接口）：
- `/api/v1/auth/login`
- `/api/v1/auth/register`
- `/api/v1/auth/captcha/send`
- `/api/v1/auth/password/forgot`
- `/api/v1/auth/refresh`

## 配置说明

### gateway.yaml

```yaml
Port: 8080                    # Gateway监听端口

Etcd:                         # etcd服务发现
  Hosts:
    - 127.0.0.1:2379

Auth:                         # JWT配置
  AccessSecret: "Skylm-im-secret-key"
  AccessExpire: 604800

WhiteList:                    # 鉴权白名单（正则）
  - ^/api/v1/auth/login$
  - ^/api/v1/auth/register$

Cors:                         # CORS配置
  AllowOrigins:
    - "http://localhost:3000"
  AllowMethods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
  AllowCredentials: true
```

## 添加新服务

以群聊服务为例：

### 1. 创建group-api服务并注册到etcd

```yaml
# app/group/api/etc/group-api.yaml
Name: group-api
Port: 10500

Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: group-api
```

### 2. （可选）添加白名单

如果有公开群聊接口：

```yaml
# gateway.yaml
WhiteList:
  - ^/api/v1/group/public/.*$
```

### 3. 启动服务

```bash
cd app/group/api && go run group-api.go
```

### 4. 自动路由

Gateway会自动识别并转发：
- `/api/v1/group/create` → group-api
- `/api/v1/group/list` → group-api

**无需修改Gateway代码！**

## 故障排查

### 1. "服务不可用"

检查后端API服务是否：
- 已启动
- 已注册到etcd
- 端口正确

```bash
# 检查etcd注册信息
etcdctl get --prefix user-api
etcdctl get --prefix friend-api
```

### 2. "鉴权失败"

检查：
- Token是否正确
- AccessSecret是否与后端服务一致
- Token是否过期

### 3. "无法解析服务名"

检查URL格式是否正确：
- ✅ `/api/v1/user/profile`
- ❌ `/user/profile`

## 性能说明

- JWT验证：<1ms
- etcd查询：<2ms（有缓存）
- 反向代理：<1ms

总体延迟：约2-3ms（可忽略）

## 架构优势

✅ 统一入口，简化前端配置  
✅ 双重验证，更安全  
✅ 服务自动发现，支持扩展  
✅ 充分利用现有API服务，无需重写逻辑  
✅ 添加新服务只需配置，无需改代码

## 后续优化

- [ ] 实现更复杂的负载均衡算法
- [ ] 添加限流功能
- [ ] 添加熔断降级
- [ ] 统一日志追踪
- [ ] 监控和指标收集
