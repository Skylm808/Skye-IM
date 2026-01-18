# Auth 模块前端对接文档

## 📋 目录

- [接口概览](#接口概览)
- [公共说明](#公共说明)
- [接口详情](#接口详情)
  - [发送验证码](#1-发送验证码)
  - [用户注册](#2-用户注册)
  - [用户登录](#3-用户登录)
  - [刷新 Token](#4-刷新-token)
  - [忘记密码](#5-忘记密码)
  - [获取用户信息](#6-获取用户信息)
  - [退出登录](#7-退出登录)
  - [修改密码](#8-修改密码)
- [错误码说明](#错误码说明)
- [前端集成示例](#前端集成示例)

---

## 接口概览

| 接口名称 | 请求方式 | 接口路径 | 是否需要认证 |
|---------|---------|---------|------------|
| 发送验证码 | POST | `/api/v1/auth/captcha/send` | ❌ |
| 用户注册 | POST | `/api/v1/auth/register` | ❌ |
| 用户登录 | POST | `/api/v1/auth/login` | ❌ |
| 刷新 Token | POST | `/api/v1/auth/refresh` | ❌ |
| 忘记密码 | POST | `/api/v1/auth/password/forgot` | ❌ |
| 获取用户信息 | GET | `/api/v1/auth/userinfo` | ✅ |
| 退出登录 | POST | `/api/v1/auth/logout` | ✅ |
| 修改密码 | POST | `/api/v1/auth/password/change` | ✅ |

---

## 公共说明

### 基础地址

```
开发环境: http://localhost:8080
生产环境: https://your-domain.com
```

### 请求头设置

#### 公开接口（无需认证）
```http
Content-Type: application/json
```

#### 需要认证的接口
```http
Content-Type: application/json
Authorization: Bearer {accessToken}
```

### 响应格式

所有接口统一返回格式：

```json
{
  "code": 0,           // 状态码，0 表示成功
  "message": "success", // 返回信息
  "data": {}           // 具体数据
}
```

### Token 机制

系统采用 **JWT 双 Token 机制**：

- **AccessToken**: 访问令牌，有效期 7 天，用于 API 认证
- **RefreshToken**: 刷新令牌，有效期 30 天，用于刷新 AccessToken

**前端存储建议**：
- AccessToken 存储在内存或 sessionStorage
- RefreshToken 存储在 httpOnly cookie 或 localStorage（加密存储）

---

## 接口详情

### 1. 发送验证码

发送邮箱验证码，用于注册或重置密码。

**接口地址**: `POST /api/v1/auth/captcha/send`

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|-------|------|-----|------|
| email | string | 是 | 邮箱地址 |
| type | string | 是 | 验证码类型：`register` 注册，`reset` 重置密码 |

**请求示例**:

```json
{
  "email": "user@example.com",
  "type": "register"
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "验证码已发送至邮箱，5分钟内有效"
  }
}
```

**注意事项**:
- ⏰ 同一邮箱 60 秒内只能发送一次验证码
- 📧 验证码有效期为 5 分钟
- 🔢 验证码为 6 位数字

---

### 2. 用户注册

使用邮箱验证码注册新账号。

**接口地址**: `POST /api/v1/auth/register`

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|-------|------|-----|------|
| username | string | 是 | 用户名，3-32 字符 |
| password | string | 是 | 密码，6-32 字符 |
| email | string | 是 | 邮箱地址 |
| captcha | string | 是 | 验证码，6 位数字 |
| phone | string | 否 | 手机号 |
| nickname | string | 否 | 昵称 |

**请求示例**:

```json
{
  "username": "skylm808",
  "password": "123456",
  "email": "user@example.com",
  "captcha": "123456",
  "phone": "13800138000",
  "nickname": "小明"
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiresIn": 604800
  }
}
```

**字段说明**:
- `accessToken`: 访问令牌
- `refreshToken`: 刷新令牌
- `expiresIn`: AccessToken 过期时间（秒），604800 = 7天

**注意事项**:
- ✅ 注册成功后自动登录，返回 Token
- 🔐 密码会使用 bcrypt 加密存储
- 📝 用户名、邮箱不能重复

---

### 3. 用户登录

使用用户名/邮箱/手机号 + 密码登录。

**接口地址**: `POST /api/v1/auth/login`

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|-------|------|-----|------|
| username | string | 是 | 用户名/邮箱/手机号 |
| password | string | 是 | 密码 |

**请求示例**:

```json
{
  "username": "skylm808",
  "password": "123456"
}
```

或使用邮箱登录：

```json
{
  "username": "user@example.com",
  "password": "123456"
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiresIn": 604800
  }
}
```

**注意事项**:
- 🔑 支持用户名、邮箱、手机号三种方式登录
- 🔒 密码错误次数过多可能触发账号锁定（待实现）

---

### 4. 刷新 Token

使用 RefreshToken 刷新 AccessToken。

**接口地址**: `POST /api/v1/auth/refresh`

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|-------|------|-----|------|
| refreshToken | string | 是 | 刷新令牌 |

**请求示例**:

```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiresIn": 604800
  }
}
```

**注意事项**:
- 🔄 建议在 AccessToken 过期前主动刷新
- 📍 前端可在请求拦截器中自动处理 Token 刷新

---

### 5. 忘记密码

通过邮箱验证码重置密码。

**接口地址**: `POST /api/v1/auth/password/forgot`

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|-------|------|-----|------|
| email | string | 是 | 注册邮箱 |
| captcha | string | 是 | 验证码，6 位数字 |
| newPassword | string | 是 | 新密码，6-32 字符 |

**请求示例**:

```json
{
  "email": "user@example.com",
  "captcha": "123456",
  "newPassword": "newpassword123"
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "密码重置成功，请使用新密码登录"
  }
}
```

**注意事项**:
- 📧 需先调用发送验证码接口，type 设为 `reset`
- 🔐 密码重置后需要重新登录

---

### 6. 获取用户信息

获取当前登录用户的信息。

**接口地址**: `GET /api/v1/auth/userinfo`

**请求头**:

```http
Authorization: Bearer {accessToken}
```

**无请求参数**

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": 1001,
    "username": "skylm808",
    "phone": "13800138000",
    "email": "user@example.com",
    "nickname": "小明",
    "avatar": "https://example.com/avatar.jpg",
    "status": 1
  }
}
```

**字段说明**:
- `id`: 用户 ID
- `status`: 用户状态（1-正常，2-禁用）

---

### 7. 退出登录

退出当前登录状态。

**接口地址**: `POST /api/v1/auth/logout`

**请求头**:

```http
Authorization: Bearer {accessToken}
```

**无请求参数**

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

**注意事项**:
- 🗑️ 前端需清除本地存储的 Token
- 🔌 需断开 WebSocket 连接

---

### 8. 修改密码

修改当前用户密码（需要旧密码验证）。

**接口地址**: `POST /api/v1/auth/password/change`

**请求头**:

```http
Authorization: Bearer {accessToken}
```

**请求参数**:

| 参数名 | 类型 | 必填 | 说明 |
|-------|------|-----|------|
| oldPassword | string | 是 | 旧密码 |
| newPassword | string | 是 | 新密码，6-32 字符 |

**请求示例**:

```json
{
  "oldPassword": "123456",
  "newPassword": "newpassword123"
}
```

**响应示例**:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "message": "密码修改成功"
  }
}
```

**注意事项**:
- 🔒 需要验证旧密码
- 🔄 修改成功后建议重新登录

---

## 错误码说明

| 错误码 | 说明 |
|-------|------|
| 0 | 成功 |
| 10001 | 参数错误 |
| 10002 | 验证码错误或已过期 |
| 10003 | 用户名已存在 |
| 10004 | 邮箱已被注册 |
| 10005 | 用户不存在 |
| 10006 | 密码错误 |
| 10007 | Token 无效或已过期 |
| 10008 | 验证码发送频繁，请稍后再试 |
| 10009 | 账号已被禁用 |
| 10010 | 旧密码错误 |

---

## 前端集成示例

### React + Axios 示例

#### 1. 创建 API 服务

```javascript
// src/services/auth.js
import axios from 'axios';

const BASE_URL = 'http://localhost:8080';

// 创建 axios 实例
const api = axios.create({
  baseURL: BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器 - 添加 Token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('accessToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// 响应拦截器 - 处理 Token 过期
api.interceptors.response.use(
  (response) => response.data,
  async (error) => {
    if (error.response?.status === 401) {
      // Token 过期，尝试刷新
      const refreshToken = localStorage.getItem('refreshToken');
      if (refreshToken) {
        try {
          const { data } = await authService.refreshToken(refreshToken);
          localStorage.setItem('accessToken', data.accessToken);
          localStorage.setItem('refreshToken', data.refreshToken);
          // 重试原请求
          error.config.headers.Authorization = `Bearer ${data.accessToken}`;
          return api.request(error.config);
        } catch (refreshError) {
          // 刷新失败，跳转登录
          localStorage.clear();
          window.location.href = '/login';
        }
      }
    }
    return Promise.reject(error);
  }
);

// Auth API 服务
export const authService = {
  // 发送验证码
  sendCaptcha: (email, type = 'register') => 
    api.post('/api/v1/auth/captcha/send', { email, type }),

  // 注册
  register: (data) => 
    api.post('/api/v1/auth/register', data),

  // 登录
  login: (username, password) => 
    api.post('/api/v1/auth/login', { username, password }),

  // 刷新 Token
  refreshToken: (refreshToken) => 
    api.post('/api/v1/auth/refresh', { refreshToken }),

  // 忘记密码
  forgotPassword: (email, captcha, newPassword) => 
    api.post('/api/v1/auth/password/forgot', { email, captcha, newPassword }),

  // 获取用户信息
  getUserInfo: () => 
    api.get('/api/v1/auth/userinfo'),

  // 退出登录
  logout: () => 
    api.post('/api/v1/auth/logout'),

  // 修改密码
  changePassword: (oldPassword, newPassword) => 
    api.post('/api/v1/auth/password/change', { oldPassword, newPassword }),
};

export default api;
```

#### 2. 使用示例

```javascript
// 登录组件示例
import { useState } from 'react';
import { authService } from '@/services/auth';

const LoginPage = () => {
  const [formData, setFormData] = useState({
    username: '',
    password: '',
  });

  const handleLogin = async () => {
    try {
      const response = await authService.login(
        formData.username,
        formData.password
      );
      
      // 存储 Token
      localStorage.setItem('accessToken', response.data.accessToken);
      localStorage.setItem('refreshToken', response.data.refreshToken);
      
      // 跳转首页
      window.location.href = '/';
    } catch (error) {
      console.error('登录失败:', error);
      alert(error.response?.data?.message || '登录失败');
    }
  };

  return (
    <div>
      <input
        type="text"
        placeholder="用户名/邮箱"
        value={formData.username}
        onChange={(e) => setFormData({ ...formData, username: e.target.value })}
      />
      <input
        type="password"
        placeholder="密码"
        value={formData.password}
        onChange={(e) => setFormData({ ...formData, password: e.target.value })}
      />
      <button onClick={handleLogin}>登录</button>
    </div>
  );
};

export default LoginPage;
```

#### 3. 注册流程示例

```javascript
const RegisterPage = () => {
  const [step, setStep] = useState(1); // 1-填写信息，2-输入验证码
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    email: '',
    captcha: '',
  });
  const [countdown, setCountdown] = useState(0);

  // 发送验证码
  const handleSendCaptcha = async () => {
    try {
      await authService.sendCaptcha(formData.email, 'register');
      alert('验证码已发送至邮箱');
      setCountdown(60);
      setStep(2);
      
      // 倒计时
      const timer = setInterval(() => {
        setCountdown((prev) => {
          if (prev <= 1) {
            clearInterval(timer);
            return 0;
          }
          return prev - 1;
        });
      }, 1000);
    } catch (error) {
      alert(error.response?.data?.message || '发送失败');
    }
  };

  // 注册
  const handleRegister = async () => {
    try {
      const response = await authService.register(formData);
      
      // 存储 Token
      localStorage.setItem('accessToken', response.data.accessToken);
      localStorage.setItem('refreshToken', response.data.refreshToken);
      
      alert('注册成功');
      window.location.href = '/';
    } catch (error) {
      alert(error.response?.data?.message || '注册失败');
    }
  };

  return (
    <div>
      {/* 表单略 */}
    </div>
  );
};
```

---

## 常见问题

### Q1: Token 存储在哪里？
**A**: 
- AccessToken 建议存储在 `sessionStorage` 或内存中
- RefreshToken 建议存储在 `localStorage`（加密后）或 httpOnly cookie

### Q2: Token 过期如何处理？
**A**: 
1. 在响应拦截器中检测 401 状态码
2. 使用 RefreshToken 调用刷新接口
3. 成功后更新 Token，重试原请求
4. 失败则清除 Token，跳转登录页

### Q3: 如何实现自动登录？
**A**: 
1. 登录成功后存储 RefreshToken
2. 页面加载时检查 Token 是否存在
3. 如果 AccessToken 过期但 RefreshToken 有效，自动刷新
4. 如果都过期则跳转登录页

### Q4: 验证码收不到怎么办？
**A**: 
1. 检查邮箱是否正确
2. 查看垃圾邮件箱
3. 确认 60 秒冷却时间已过
4. 联系管理员检查邮件服务配置

---

## 更新日志

| 版本 | 日期 | 更新内容 |
|------|------|---------|
| v1.0 | 2026-01-13 | 初始版本，包含所有基础认证接口 |

---

**文档维护**: Skylm  
**最后更新**: 2026-01-13
