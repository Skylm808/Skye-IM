# Gateway 网关前端对接文档

##📋 目录

- [概述](#概述)
- [访问地址](#访问地址)
- [路由规则](#路由规则)
- [鉴权机制](#鉴权机制)
- [CORS 配置](#cors-配置)
- [错误处理](#错误处理)
- [常见问题](#常见问题)

---

## 概述

Gateway 是 SkyeIM 的统一 API 网关，所有前端请求都应该通过 Gateway 访问后端服务。

### 核心功能

| 功能 | 说明 |
|------|------|
| 统一入口 | 所有请求通过 8080 端口访问 |
| JWT 鉴权 | 自动验证 Token 有效性 |
| 服务发现 | 基于 etcd 自动路由到后端服务 |
| 反向代理 | 透明转发请求和响应 |
| CORS 支持 | 支持跨域请求 |

### 技术架构

```
前端应用 (localhost:3000)
    ↓ HTTP 请求
Gateway (localhost:8080)
    ↓ 服务发现 (etcd)
    ↓ JWT 鉴权
    ↓ 反向代理
后端服务 (10001/10100/10200...)
```

---

## 访问地址

### 开发环境

```
Gateway: http://localhost:8080
```

### 生产环境

```
Gateway: https://your-domain.com
```

**重要提示**：
- ✅ 前端只需要配置 Gateway 地址
- ❌ 不要直接访问后端服务（10001/10100 等端口）
- 🔐 所有请求自动经过 Gateway 鉴权和转发

---

## 路由规则

Gateway 根据 URL 路径自动识别目标服务。

### URL 格式

```
统一格式: http://localhost:8080/api/v1/{service}/{endpoint}
```

### 路由映射表

| URL 模式 | 目标服务 | 实际端口 | 说明 |
|----------|---------|---------|------|
| `/api/v1/auth/*` | auth-api | 10001 | 认证服务 |
| `/api/v1/user/*` | user-api | 10100 | 用户服务 |
| `/api/v1/friend/*` | friend-api | 10200 | 好友服务 |
| `/api/v1/message/*` | message-api | 10400 | 消息服务 |
| `/api/v1/group/*` | group-api | 10500 | 群组服务 |
| `/api/v1/upload/*` | upload-api | 10600 | 上传服务 |

### 路由示例

```javascript
// ✅ 正确：通过 Gateway 访问
fetch('http://localhost:8080/api/v1/user/profile', {
  headers: {
    'Authorization': 'Bearer ' + token
  }
})

// ❌ 错误：直接访问后端服务
fetch('http://localhost:10100/api/v1/user/profile')

// ❌ 错误：URL 格式不正确
fetch('http://localhost:8080/user/profile')
```

---

## 鉴权机制

### 白名单接口（无需 Token）

以下接口**不需要**在请求头中携带 Token：

| 接口路径 | 说明 |
|---------|------|
| `/api/v1/auth/login` | 用户登录 |
| `/api/v1/auth/register` | 用户注册 |
| `/api/v1/auth/captcha/send` | 发送验证码 |
| `/api/v1/auth/password/forgot` | 忘记密码 |
| `/api/v1/auth/refresh` | 刷新 Token |

**请求示例**（无需 Token）：

```javascript
fetch('http://localhost:8080/api/v1/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    username: 'test',
    password: '123456'
  })
})
```

### 需要鉴权的接口

除白名单外的所有接口都需要在请求头中携带 **AccessToken**。

**请求头格式**：

```http
Authorization: Bearer {accessToken}
Content-Type: application/json
```

**请求示例**（需要 Token）：

```javascript
const token = localStorage.getItem('accessToken');

fetch('http://localhost:8080/api/v1/user/profile', {
  method: 'GET',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
})
```

### Token 自动注入

Gateway 验证 Token 后，会自动将用户信息注入到请求头：

| Header | 说明 | 示例值 |
|--------|------|-------|
| `X-User-Id` | 用户 ID | `1001` |
| `X-Username` | 用户名 | `skylm808` |

**用途**：后端服务可以直接从 Header 获取用户信息，无需再次解析 JWT。

---

## CORS 配置

### 允许的来源

Gateway 默认允许以下来源的跨域请求：

```
http://localhost:3000   # React 默认端口
http://localhost:5173   # Vite 默认端口
http://localhost:5174   # Vite 备用端口
```

### 允许的 HTTP 方法

```
GET
POST
PUT
DELETE
OPTIONS
```

### 允许的请求头

```
Content-Type
Authorization
X-Requested-With
```

### 前端配置注意事项

**1. 携带 Cookie（如果需要）**：

```javascript
fetch('http://localhost:8080/api/v1/user/profile', {
  credentials: 'include',  // 允许发送 Cookie
  headers: {
    'Authorization': `Bearer ${token}`
  }
})
```

**2. 自定义请求头**：

所有自定义请求头都会被 Gateway 允许，无需额外配置。

**3. 预检请求（OPTIONS）**：

Gateway 会自动处理 `OPTIONS` 请求，前端无需关心。

---

## 错误处理

### HTTP 状态码

| 状态码 | 说明 | 常见原因 |
|-------|------|---------|
| 200 | 成功 | 请求正常处理 |
| 400 | 参数错误 | URL 格式错误、无法解析服务名 |
| 401 | 鉴权失败 | Token 缺失、无效或过期 |
| 502 | 网关错误 | 后端服务不可用 |
| 503 | 服务不可用 | 服务未启动或未注册到 etcd |

### 错误响应格式

```json
{
  "error": "错误描述信息"
}
```

### 常见错误及解决方案

#### 1. 401 Unauthorized - 鉴权失败

**错误信息**：
```
鉴权失败: Token无效
鉴权失败: 缺少Authorization header
```

**原因**：
- Token 过期
- Token 格式错误
- 未在请求头中携带 Token
- AccessSecret 不一致

**解决方案**：
```javascript
// 1. 检查 Token 是否存在
const token = localStorage.getItem('accessToken');
if (!token) {
  // 跳转登录页
  window.location.href = '/login';
}

// 2. 检查 Token 格式
// 正确: Bearer eyJhbGciOiJIUzI1NiIs...
// 错误: eyJhbGciOiJIUzI1NiIs... (缺少 Bearer 前缀)

// 3. Token 过期，使用 RefreshToken 刷新
if (response.status === 401) {
  const refreshToken = localStorage.getItem('refreshToken');
  const newToken = await refreshAccessToken(refreshToken);
  // 重试原请求
}
```

#### 2. 503 Service Unavailable - 服务不可用

**错误信息**：
```
服务不可用: 服务 user-api 无可用实例
```

**原因**：
- 后端服务未启动
- 后端服务未注册到 etcd
- etcd 服务未启动

**解决方案**：
1. 检查后端服务是否运行
2. 检查后端服务的 etcd 配置
3. 联系后端开发人员

#### 3. 502 Bad Gateway - 网关错误

**错误信息**：
```
后端服务不可用: dial tcp 127.0.0.1:10100: connect: connection refused
```

**原因**：
- 后端服务突然崩溃
- 端口配置错误

**解决方案**：
1. 检查后端服务日志
2. 重启后端服务
3. 联系后端开发人员

#### 4. 400 Bad Request - 无法解析服务名

**错误信息**：
```
无法解析服务名
```

**原因**：
- URL 格式不正确

**解决方案**：
```javascript
// ✅ 正确格式
/api/v1/user/profile
/api/v1/friend/list

// ❌ 错误格式
/user/profile          // 缺少 /api/v1
/api/user/profile      // 缺少版本号 v1
/api/v1//profile       // 多余的斜杠
```

---

## 常见问题

### Q1: 前端需要配置多个后端服务地址吗？

**A**: 不需要。前端只需要配置一个 Gateway 地址（`http://localhost:8080`），Gateway 会自动路由到对应的后端服务。

```javascript
// ✅ 推荐：统一配置
const API_BASE_URL = 'http://localhost:8080';

// ❌ 不推荐：配置多个地址
const AUTH_URL = 'http://localhost:10001';
const USER_URL = 'http://localhost:10100';
const FRIEND_URL = 'http://localhost:10200';
```

---

### Q2: Token 存储在哪里？

**A**: 
- **AccessToken**: 建议存储在 `sessionStorage` 或内存中（安全性更高）
- **RefreshToken**: 建议存储在 `localStorage`

```javascript
// 登录成功后
localStorage.setItem('accessToken', response.accessToken);
localStorage.setItem('refreshToken', response.refreshToken);

// 发送请求时
const token = localStorage.getItem('accessToken');
headers.Authorization = `Bearer ${token}`;
```

---

### Q3: 如何处理 Token 过期？

**A**: 在 Axios 响应拦截器中统一处理：

```javascript
axios.interceptors.response.use(
  response => response,
  async error => {
    if (error.response?.status === 401) {
      // Token 过期，尝试刷新
      const refreshToken = localStorage.getItem('refreshToken');
      
      try {
        const { data } = await axios.post('/api/v1/auth/refresh', {
          refreshToken
        });
        
        // 更新 Token
        localStorage.setItem('accessToken', data.accessToken);
        localStorage.setItem('refreshToken', data.refreshToken);
        
        // 重试原请求
        error.config.headers.Authorization = `Bearer ${data.accessToken}`;
        return axios.request(error.config);
      } catch (refreshError) {
        // 刷新失败，跳转登录页
        localStorage.clear();
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);
```

---

### Q4: Gateway 和后端服务都会验证 Token 吗？

**A**: 是的，采用**双重验证**机制：

1. **Gateway 层**：验证 Token 有效性，注入用户信息
2. **后端服务层**：再次验证 Token（可选，但推荐保留）

**优点**：
- 更安全：即使 Gateway 被绕过，后端仍有保护
- 灵活：后端服务可以独立运行（跳过 Gateway）

---

### Q5: 开发环境如何跳过 Gateway？

**A**: 开发时可以直接访问后端服务，但需要自己处理 CORS 和鉴权：

```javascript
// 开发环境：直接访问后端（需要后端设置 CORS）
const API_BASE_URL = process.env.NODE_ENV === 'development'
  ? 'http://localhost:10100'  // 直接访问后端
  : 'http://localhost:8080';  // 生产通过 Gateway

// 推荐：始终通过 Gateway（一致性更好）
const API_BASE_URL = 'http://localhost:8080';
```

---

### Q6: Gateway 是否支持 WebSocket？

**A**: 当前 Gateway 主要处理 HTTP 请求。WebSocket 连接应该**直接连接** WebSocket 服务（`ws://localhost:10300`），不经过 Gateway。

```javascript
// HTTP 请求：通过 Gateway
axios.post('http://localhost:8080/api/v1/message/send', data);

// WebSocket 连接：直接连接
const ws = new WebSocket('ws://localhost:10300/ws?token=' + token);
```

---

### Q7: 如何判断请求是否成功？

**A**: 
1. HTTP 状态码为 `200`
2. 响应体中 `code` 为 `0`（根据后端约定）

```javascript
const response = await fetch('http://localhost:8080/api/v1/user/profile', {
  headers: { 'Authorization': `Bearer ${token}` }
});

if (response.ok) {  // HTTP 200-299
  const data = await response.json();
  if (data.code === 0) {  // 业务成功
    console.log('用户信息:', data.data);
  } else {
    console.error('业务错误:', data.message);
  }
} else {
  console.error('HTTP 错误:', response.status);
}
```

---

### Q8: Gateway 添加新服务需要前端改代码吗？

**A**: 不需要。Gateway 会自动识别新服务。

```
后端新增服务: group-api (端口 10500)

前端直接访问:
http://localhost:8080/api/v1/group/create

无需修改前端配置！
```

---

## 性能说明

Gateway 对请求的影响：

| 操作 | 耗时 |
|------|------|
| JWT 验证 | < 1 ms |
| etcd 查询 | < 2 ms（有缓存） |
| 反向代理 | < 1 ms |
| **总计** | **约 2-3 ms** |

对用户几乎无感知，可以忽略。

---

## 监控与日志

### Gateway 日志

Gateway 会记录所有请求的详细信息：

```
请求完成: POST /api/v1/auth/login → auth-api, 耗时: 45ms
转发请求: GET /api/v1/user/profile → user-api (127.0.0.1:10100)
鉴权失败: path=/api/v1/friend/list, err=Token无效
```

### 健康检查

Gateway 本身不提供健康检查接口，但可以通过访问任意白名单接口验证：

```bash
curl http://localhost:8080/api/v1/auth/login
# 返回 400 或其他错误码表示 Gateway 正常运行
```

---

## 更新日志

| 版本 | 日期 | 更新内容 |
|------|------|---------|
| v1.0 | 2026-01-13 | 初始版本，支持基础路由和鉴权 |

---

**文档维护**: Skylm  
**最后更新**: 2026-01-13  
**相关文档**: [Gateway 架构设计](../ARCHITECTURE.md)
