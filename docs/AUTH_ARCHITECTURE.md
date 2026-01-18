# Auth 微服务架构与流程讲解

## 一、服务概述

Auth 服务是 myIM 即时通讯系统的**用户认证微服务**，负责用户的注册、登录、Token 管理等核心认证功能。

### 技术栈

| 组件  | 技术选型    | 用途             |
| --- | ------- | -------------- |
| 框架  | go-zero | 微服务框架          |
| 数据库 | MySQL   | 用户数据持久化        |
| 缓存  | Redis   | 验证码存储、Model 缓存 |
| 认证  | JWT     | 无状态 Token 认证   |
| 密码  | bcrypt  | 密码加密           |
| 邮件  | gomail  | 发送验证码          |

---

## 二、目录结构

```
app/auth/
├── auth.go                 # 服务入口
├── auth.api                # API 定义文件
├── etc/
│   └── auth-api.yaml       # 服务配置
├── internal/
│   ├── config/
│   │   └── config.go       # 配置结构体
│   ├── handler/            # HTTP 处理器（自动生成）
│   │   ├── routes.go       # 路由注册
│   │   ├── public/         # 公开接口处理器
│   │   └── user/           # 需认证接口处理器
│   ├── logic/              # 业务逻辑层
│   │   ├── public/         # 公开接口逻辑
│   │   │   ├── registerlogic.go
│   │   │   ├── loginlogic.go
│   │   │   ├── sendcaptchalogic.go
│   │   │   └── refreshtokenlogic.go
│   │   └── user/           # 需认证接口逻辑
│   │       ├── getuserinfologic.go
│   │       └── logoutlogic.go
│   ├── svc/
│   │   └── serviceContext.go  # 服务上下文（依赖注入）
│   └── types/
│       └── types.go        # 请求/响应类型（自动生成）
└── model/
    ├── user.sql            # 数据库 DDL
    ├── usermodel.go        # Model 接口（可扩展）
    ├── userModel_gen.go    # Model 实现（goctl 生成，带缓存）
    └── vars.go             # 错误定义
```

---

## 三、核心流程

### 3.1 发送验证码流程

```
┌─────────┐     POST /captcha/send      ┌─────────────┐
│  前端   │ ─────────────────────────► │  Handler    │
└─────────┘                             └──────┬──────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │   Logic     │
                                        │ SendCaptcha │
                                        └──────┬──────┘
                                               │
                    ┌──────────────────────────┼──────────────────────────┐
                    │                          │                          │
                    ▼                          ▼                          ▼
             ┌─────────────┐           ┌─────────────┐           ┌─────────────┐
             │ 检查邮箱    │           │ 检查发送    │           │ 生成验证码  │
             │ 是否已注册  │           │ 频率限制    │           │ (6位随机数) │
             └──────┬──────┘           └──────┬──────┘           └──────┬──────┘
                    │                          │                          │
                    ▼                          ▼                          ▼
             ┌─────────────┐           ┌─────────────┐           ┌─────────────┐
             │   MySQL     │           │   Redis     │           │  QQ邮箱     │
             │  查询用户   │           │ 60秒限制    │           │  SMTP发送   │
             └─────────────┘           └─────────────┘           └──────┬──────┘
                                                                        │
                                                                        ▼
                                                                 ┌─────────────┐
                                                                 │   Redis     │
                                                                 │ 存储验证码  │
                                                                 │ (5分钟过期) │
                                                                 └─────────────┘
```

### 3.2 用户注册流程

```
┌─────────┐     POST /register          ┌─────────────┐
│  前端   │ ─────────────────────────► │  Handler    │
│         │  {email, captcha,          └──────┬──────┘
│         │   username, password}             │
└─────────┘                                   ▼
                                        ┌─────────────┐
                                        │   Logic     │
                                        │  Register   │
                                        └──────┬──────┘
                                               │
        ┌──────────────────────────────────────┼──────────────────────────────────────┐
        │                                      │                                      │
        ▼                                      ▼                                      ▼
 ┌─────────────┐                       ┌─────────────┐                       ┌─────────────┐
 │ 验证验证码  │                       │ 检查用户名  │                       │ 检查邮箱    │
 │ (Redis)     │                       │ 是否存在    │                       │ 是否存在    │
 └──────┬──────┘                       └──────┬──────┘                       └──────┬──────┘
        │                                      │                                      │
        │ 验证成功                             │ 不存在                               │ 不存在
        ▼                                      ▼                                      ▼
 ┌─────────────┐                       ┌─────────────┐                       ┌─────────────┐
 │ 删除验证码  │                       │ bcrypt 加密 │                       │ 插入用户    │
 │ (防重复用)  │                       │ 密码        │                       │ (MySQL)     │
 └─────────────┘                       └─────────────┘                       └──────┬──────┘
                                                                                    │
                                                                                    ▼
                                                                             ┌─────────────┐
                                                                             │ 生成 JWT    │
                                                                             │ Token Pair  │
                                                                             └──────┬──────┘
                                                                                    │
                                                                                    ▼
                                                                             ┌─────────────┐
                                                                             │ 返回 Token  │
                                                                             │ 给前端      │
                                                                             └─────────────┘
```

### 3.3 用户登录流程

```
┌─────────┐     POST /login             ┌─────────────┐
│  前端   │ ─────────────────────────► │  Handler    │
│         │  {username, password}       └──────┬──────┘
└─────────┘                                    │
                                               ▼
                                        ┌─────────────┐
                                        │   Logic     │
                                        │   Login     │
                                        └──────┬──────┘
                                               │
                    ┌──────────────────────────┴──────────────────────────┐
                    │                                                      │
                    ▼                                                      ▼
             ┌─────────────┐                                        ┌─────────────┐
             │ 查询用户    │                                        │ 支持多种    │
             │ (优先缓存)  │                                        │ 登录方式    │
             └──────┬──────┘                                        └─────────────┘
                    │                                                      │
                    │  ┌────────────────────────────────────────────────────┘
                    │  │
                    ▼  ▼
             ┌─────────────┐
             │ 用户名/手机 │
             │ /邮箱 登录  │
             └──────┬──────┘
                    │
                    ▼
             ┌─────────────┐     不匹配     ┌─────────────┐
             │ bcrypt 验证 │ ─────────────► │ 返回错误    │
             │ 密码        │                └─────────────┘
             └──────┬──────┘
                    │ 匹配
                    ▼
             ┌─────────────┐
             │ 检查用户    │
             │ 状态        │
             └──────┬──────┘
                    │ 正常
                    ▼
             ┌─────────────┐
             │ 生成 JWT    │
             │ Token Pair  │
             └──────┬──────┘
                    │
                    ▼
             ┌─────────────┐
             │ 返回 Token  │
             └─────────────┘
```

### 3.4 Token 刷新流程

```
┌─────────┐     POST /refresh           ┌─────────────┐
│  前端   │ ─────────────────────────► │  Handler    │
│         │  {refreshToken}             └──────┬──────┘
└─────────┘                                    │
                                               ▼
                                        ┌─────────────┐
                                        │ 解析并验证  │
                                        │ RefreshToken│
                                        └──────┬──────┘
                                               │
                    ┌──────────────────────────┴──────────────────────────┐
                    │                                                      │
                    ▼                                                      ▼
             ┌─────────────┐                                        ┌─────────────┐
             │ 验证 Token  │                                        │ 检查用户    │
             │ 类型        │                                        │ 是否存在    │
             └──────┬──────┘                                        └──────┬──────┘
                    │                                                      │
                    │ 是 RefreshToken                                      │ 存在且正常
                    ▼                                                      ▼
             ┌─────────────────────────────────────────────────────────────┐
             │                    生成新的 Token Pair                       │
             │              (AccessToken + RefreshToken)                   │
             └─────────────────────────────────────────────────────────────┘
```

---

## 四、核心组件详解

### 4.1 JWT 双 Token 机制

```go
// AccessToken: 短期有效（7天），用于 API 认证
// RefreshToken: 长期有效（30天），用于刷新 AccessToken

type TokenPair struct {
    AccessToken  string  // 访问令牌
    RefreshToken string  // 刷新令牌
    ExpiresIn    int64   // 过期时间（秒）
}

type Claims struct {
    UserId    int64   // 用户ID
    Username  string  // 用户名
    TokenType string  // "access" 或 "refresh"
    jwt.RegisteredClaims
}
```

**为什么使用双 Token？**

- AccessToken 短期有效，降低泄露风险
- RefreshToken 长期有效，提升用户体验（无需频繁登录）
- 分离职责：AccessToken 用于 API 访问，RefreshToken 仅用于续期

### 4.2 Model 缓存机制

go-zero 的 `goctl model` 生成的代码自带 Redis 缓存：

```go
// 缓存 Key 前缀
cacheUserIdPrefix       = "cache:user:id:"
cacheUserEmailPrefix    = "cache:user:email:"
cacheUserPhonePrefix    = "cache:user:phone:"
cacheUserUsernamePrefix = "cache:user:username:"

// 查询时自动使用缓存
func (m *defaultUserModel) FindOne(ctx context.Context, id uint64) (*User, error) {
    userIdKey := fmt.Sprintf("%s%v", cacheUserIdPrefix, id)
    // 先查缓存，没有再查数据库
    err := m.QueryRowCtx(ctx, &resp, userIdKey, func(...) error {
        // 查询数据库
        return conn.QueryRowCtx(ctx, v, query, id)
    })
    // ...
}

// 增删改时自动清除缓存
func (m *defaultUserModel) Update(ctx context.Context, newData *User) error {
    // 清除所有相关缓存 Key
    _, err = m.ExecCtx(ctx, func(...) {
        // 执行更新
    }, userEmailKey, userIdKey, userPhoneKey, userUsernameKey)
}
```

### 4.3 验证码服务

```go
// Redis Key 设计
captchaKeyPrefix   = "captcha:email:"   // 验证码存储
sendLimitKeyPrefix = "captcha:limit:"   // 发送频率限制

// 功能
- Generate()      生成6位随机验证码
- Store()         存储验证码（5分钟过期）
- Verify()        验证并删除验证码（防重复使用）
- CheckSendLimit() 检查60秒发送限制
- SetSendLimit()  设置发送限制
```

### 4.4 服务上下文（依赖注入）

```go
type ServiceContext struct {
    Config         config.Config       // 配置
    UserModel      model.UserModel     // 用户数据模型
    Validator      *validator.Validate // 参数校验器
    EmailSender    *email.Sender       // 邮件发送器
    CaptchaService *captcha.Service    // 验证码服务
}
```

所有 Logic 通过 ServiceContext 获取依赖，实现了：

- **解耦**：业务逻辑不直接依赖具体实现
- **可测试**：可以 Mock 依赖进行单元测试
- **统一管理**：所有依赖在一处初始化

---

## 五、安全设计

### 5.1 密码安全

```go
// 使用 bcrypt 加密，自动加盐
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// 验证时比较 hash
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

### 5.2 验证码安全

- **频率限制**：60秒内只能发送一次
- **有效期**：5分钟后自动过期
- **一次性**：验证成功后立即删除
- **邮箱验证**：确保邮箱真实有效

### 5.3 JWT 安全

- **Secret 长度**：≥8 字符（go-zero 强制要求）
- **Token 类型**：区分 Access 和 Refresh，防止混用
- **用户状态检查**：刷新 Token 时检查用户是否被禁用

---

## 六、API 接口

| 接口                          | 方法   | 认证  | 描述       |
| --------------------------- | ---- | --- | -------- |
| `/api/v1/auth/captcha/send` | POST | 否   | 发送邮箱验证码  |
| `/api/v1/auth/register`     | POST | 否   | 用户注册     |
| `/api/v1/auth/login`        | POST | 否   | 用户登录     |
| `/api/v1/auth/refresh`      | POST | 否   | 刷新 Token |
| `/api/v1/auth/logout`       | POST | 是   | 退出登录     |
| `/api/v1/auth/userinfo`     | GET  | 是   | 获取用户信息   |

---

## 七、配置说明

```yaml
# auth-api.yaml

Name: auth-api
Host: 0.0.0.0
Port: 8888

MySQL:
  DataSource: root:password@tcp(127.0.0.1:3306)/im_auth?charset=utf8mb4&parseTime=True&loc=Local

Cache:                    # Model 缓存
  - Host: 127.0.0.1:6379
    Type: node
    Pass: "123456"

Redis:                    # 验证码等业务缓存
  Host: 127.0.0.1:6379
  Type: node
  Pass: "123456"

Auth:                     # JWT 配置
  AccessSecret: "your-secret-key"
  AccessExpire: 604800    # 7天

RefreshToken:
  Secret: "your-refresh-secret"
  Expire: 2592000         # 30天

Email:                    # QQ 邮箱 SMTP
  Host: smtp.qq.com
  Port: 465
  Username: your@qq.com
  Password: your-auth-code
  From: "myIM系统"

Captcha:
  Expire: 300             # 5分钟
  Length: 6               # 6位
```

---

## 八、总结

Auth 微服务采用**分层架构**，遵循 go-zero 的最佳实践：

1. **Handler 层**：处理 HTTP 请求，参数绑定
2. **Logic 层**：核心业务逻辑
3. **Model 层**：数据访问，带缓存
4. **ServiceContext**：依赖注入容器

通过 `goctl` 工具生成代码骨架，开发者只需专注于 Logic 层的业务实现，大大提高了开发效率。
