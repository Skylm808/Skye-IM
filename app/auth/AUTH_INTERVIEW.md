# Auth 微服务面试问答

## 一、项目架构类

### Q1: 为什么选择 go-zero 框架？

**答：**

go-zero 是一个集成了各种工程实践的 Web 和 RPC 框架，选择它的原因：

1. **高性能**：内置连接池、缓存、熔断等，QPS 可达百万级
2. **代码生成**：`goctl` 工具可以从 API 文件自动生成代码骨架，提高开发效率
3. **微服务友好**：原生支持 gRPC，服务发现，负载均衡
4. **内置最佳实践**：自带限流、熔断、降级、超时控制
5. **Model 缓存**：自动生成带 Redis 缓存的数据访问层

### Q2: Auth 服务的整体架构是怎样的？

**答：**

采用经典的**三层架构**：

```
┌─────────────────────────────────────┐
│           Handler 层                │  ← HTTP 请求处理、参数绑定
├─────────────────────────────────────┤
│           Logic 层                  │  ← 核心业务逻辑
├─────────────────────────────────────┤
│           Model 层                  │  ← 数据访问、缓存
├─────────────────────────────────────┤
│     MySQL        Redis              │  ← 持久化存储
└─────────────────────────────────────┘
```

通过 **ServiceContext** 实现依赖注入，所有依赖统一管理。

### Q3: 为什么要把接口分成 public 和 user 两个 group？

**答：**

这是基于**认证需求**的划分：

- **public 组**：无需认证的接口（注册、登录、发送验证码、刷新 Token）
- **user 组**：需要 JWT 认证的接口（获取用户信息、退出登录）

在 `.api` 文件中通过 `@server` 注解配置：

```go
@server (
    prefix: /api/v1/auth
    group:  user
    jwt:    Auth          // 启用 JWT 认证
)
```

好处：
1. **代码组织清晰**：不同权限的接口分开管理
2. **自动路由分组**：goctl 会生成对应目录结构
3. **中间件复用**：可以对 group 统一添加中间件

---

## 二、认证安全类

### Q4: 为什么使用双 Token（AccessToken + RefreshToken）机制？

**答：**

单 Token 的问题：
- 有效期短 → 用户频繁登录，体验差
- 有效期长 → 泄露风险大，安全性差

双 Token 方案：

| Token | 有效期 | 用途 | 存储位置 |
|-------|--------|------|----------|
| AccessToken | 7天（短） | API 认证 | 内存/localStorage |
| RefreshToken | 30天（长） | 刷新 AccessToken | httpOnly Cookie（更安全）|

**工作流程：**
1. 登录成功 → 返回双 Token
2. API 请求 → 携带 AccessToken
3. AccessToken 过期 → 用 RefreshToken 换取新的双 Token
4. RefreshToken 过期 → 重新登录

**安全优势：**
- AccessToken 泄露影响有限（短期有效）
- RefreshToken 仅用于刷新，减少暴露机会
- 可以实现**无感刷新**，提升用户体验

### Q5: 密码为什么使用 bcrypt 而不是 MD5/SHA256？

**答：**

**MD5/SHA256 的问题：**
1. 计算速度太快 → 容易被暴力破解
2. 需要手动加盐 → 容易实现不当
3. 彩虹表攻击 → 预计算的哈希表可以快速破解

**bcrypt 的优势：**
1. **自动加盐**：每次生成不同的盐值
2. **慢哈希**：故意设计得很慢（可调节 cost），增加破解成本
3. **抗 GPU 攻击**：内存密集型，GPU 难以加速
4. **成熟可靠**：业界标准，经过长期验证

```go
// bcrypt 使用示例
hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
// DefaultCost = 10，每增加1，计算时间翻倍
```

### Q6: 验证码系统是如何防止滥用的？

**答：**

实现了多层防护：

1. **发送频率限制**：同一邮箱 60 秒内只能发送一次
   ```go
   key := "captcha:limit:" + email
   redis.SetexCtx(ctx, key, "1", 60)
   ```

2. **有效期控制**：验证码 5 分钟后自动过期
   ```go
   redis.SetexCtx(ctx, key, code, 300)
   ```

3. **一次性使用**：验证成功后立即删除
   ```go
   redis.DelCtx(ctx, key)
   ```

4. **邮箱验证**：确保邮箱真实有效，防止虚假注册

5. **可扩展**：可以增加图形验证码、IP 限制等

### Q7: JWT Token 是如何验证的？go-zero 中间件如何工作？

**答：**

go-zero 的 JWT 中间件工作流程：

```
HTTP 请求 → 提取 Authorization Header → 解析 Token → 验证签名和过期时间 → 注入 Context
```

1. **提取 Token**：从 `Authorization: Bearer <token>` 提取
2. **解析验证**：使用配置的 Secret 验证签名
3. **注入 Context**：将 Claims 信息注入到请求上下文

```go
// 在 Logic 中获取用户信息
userId := l.ctx.Value("userId")  // 从 Context 获取
```

配置方式：
```yaml
Auth:
  AccessSecret: "your-secret-key"  # ≥8 字符
  AccessExpire: 604800
```

---

## 三、缓存与数据库类

### Q8: goctl 生成的 Model 缓存是如何工作的？

**答：**

**缓存策略：Cache-Aside（旁路缓存）**

**读操作：**
```
查询 → 先查 Redis → 命中则返回
                  → 未命中 → 查 MySQL → 写入 Redis → 返回
```

**写操作：**
```
更新/删除 → 操作 MySQL → 删除相关 Redis 缓存
```

**缓存 Key 设计：**
```go
cacheUserIdPrefix       = "cache:user:id:"        // 主键缓存
cacheUserEmailPrefix    = "cache:user:email:"     // 唯一索引缓存
cacheUserPhonePrefix    = "cache:user:phone:"
cacheUserUsernamePrefix = "cache:user:username:"
```

**自动清理：** 当数据变更时，自动删除所有相关的缓存 Key，保证一致性。

### Q9: 缓存和数据库的一致性如何保证？

**答：**

go-zero 采用的是**先更新数据库，再删除缓存**的策略：

```go
func (m *defaultUserModel) Update(ctx context.Context, newData *User) error {
    // 1. 查询原数据（获取缓存 Key）
    data, err := m.FindOne(ctx, newData.Id)
    
    // 2. 执行数据库更新
    _, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) {
        return conn.ExecCtx(ctx, query, ...)
    }, 
    // 3. 删除所有相关缓存
    userEmailKey, userIdKey, userPhoneKey, userUsernameKey)
}
```

**为什么不是"先删缓存，再更新数据库"？**

在并发场景下：
```
线程A: 删除缓存
线程B: 读取（缓存未命中）→ 读数据库旧值 → 写入缓存
线程A: 更新数据库
结果: 缓存是旧值，数据库是新值 → 不一致！
```

**为什么"先更新数据库，再删缓存"更好？**

即使出现并发：
```
线程A: 更新数据库
线程B: 读取（缓存命中）→ 返回旧值（短暂不一致）
线程A: 删除缓存
下次读取: 缓存未命中 → 读数据库新值 → 最终一致
```

最终一致性可以保证，短暂的不一致是可接受的。

### Q10: 如果 Redis 挂了，服务会怎样？

**答：**

go-zero 的缓存层设计了**降级机制**：

1. **缓存穿透保护**：即使 Redis 不可用，会直接查询 MySQL
2. **错误处理**：缓存操作失败不会导致整个请求失败
3. **熔断机制**：go-zero 内置熔断，Redis 频繁失败时会自动降级

```go
err := m.QueryRowCtx(ctx, &resp, cacheKey, func(...) error {
    // 如果 Redis 失败，这个回调函数会被执行，直接查 MySQL
    return conn.QueryRowCtx(ctx, v, query, id)
})
```

**建议的生产实践：**
- Redis 主从/集群部署
- 监控 Redis 健康状态
- 设置合理的超时时间
- 关键数据持久化

---

## 四、性能与优化类

### Q11: 如何优化登录接口的性能？

**答：**

当前实现的优化点：

1. **Model 缓存**：热点用户数据缓存在 Redis
2. **多方式登录**：用户名/手机/邮箱查询走不同的索引

进一步优化建议：

1. **连接池**：go-zero 默认使用连接池
2. **异步处理**：登录日志等非核心逻辑异步化
3. **限流**：防止暴力破解
4. **布隆过滤器**：快速判断用户是否存在

```go
// 登录查询优化：利用唯一索引
user, err = l.svcCtx.UserModel.FindOneByUsername(ctx, username)  // 走 idx_username
user, err = l.svcCtx.UserModel.FindOneByPhone(ctx, phone)        // 走 idx_phone
user, err = l.svcCtx.UserModel.FindOneByEmail(ctx, email)        // 走 idx_email
```

### Q12: 发送邮件是同步的，会不会阻塞请求？

**答：**

**当前实现**：同步发送，确实会阻塞（1-2秒）

**优化方案：异步发送**

```go
// 方案1：使用 goroutine
go func() {
    if err := emailSender.SendCode(email, code); err != nil {
        logx.Errorf("发送邮件失败: %v", err)
    }
}()

// 方案2：使用消息队列（推荐生产环境）
// 1. 生成验证码 → 存 Redis
// 2. 发送消息到 MQ（邮件任务）
// 3. 立即返回"验证码已发送"
// 4. 消费者异步发送邮件
```

**消息队列的好处：**
- 削峰填谷
- 失败重试
- 解耦服务

---

## 五、问题排查类

### Q13: 如果用户反馈"验证码错误"，如何排查？

**答：**

排查步骤：

1. **查看后端日志**
   ```bash
   # 搜索验证相关日志
   grep "验证码" logs/auth-api.log
   ```

2. **检查 Redis 中的验证码**
   ```bash
   redis-cli GET "captcha:email:user@example.com"
   redis-cli TTL "captcha:email:user@example.com"
   ```

3. **常见原因**
   - 用户输入了旧的验证码（发送了多次）
   - 验证码已过期（超过5分钟）
   - 邮箱大小写不一致（代码中已做 ToLower 处理）
   - Redis 连接问题

4. **添加调试日志**
   ```go
   l.Logger.Infof("验证验证码: email=%s, input=%s, stored=%s", 
       email, inputCode, storedCode)
   ```

### Q14: 线上出现"NOAUTH Authentication required"错误怎么办？

**答：**

这是 **Redis 需要密码但配置中没有提供**的错误。

**排查步骤：**

1. **确认 Redis 是否需要密码**
   ```bash
   redis-cli CONFIG GET requirepass
   ```

2. **检查配置文件**
   ```yaml
   Redis:
     Host: 127.0.0.1:6379
     Type: node
     Pass: "your-password"  # 确保密码正确
   ```

3. **检查是否有多处 Redis 配置**
   ```yaml
   Cache:    # Model 缓存的 Redis
     Pass: "123456"
   Redis:    # 业务的 Redis
     Pass: "123456"
   ```

---

## 六、扩展设计类

### Q15: 如果要支持第三方登录（微信/GitHub），如何设计？

**答：**

**数据库设计**：新增用户第三方绑定表

```sql
CREATE TABLE `user_oauth` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `user_id` bigint NOT NULL COMMENT '用户ID',
    `provider` varchar(32) NOT NULL COMMENT '第三方平台: wechat/github/google',
    `open_id` varchar(128) NOT NULL COMMENT '第三方用户ID',
    `union_id` varchar(128) DEFAULT NULL COMMENT '微信 UnionID',
    `access_token` varchar(512) DEFAULT NULL,
    `refresh_token` varchar(512) DEFAULT NULL,
    `expires_at` datetime DEFAULT NULL,
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_provider_openid` (`provider`, `open_id`),
    KEY `idx_user_id` (`user_id`)
);
```

**接口设计**：

```
POST /api/v1/auth/oauth/github     # GitHub 登录
POST /api/v1/auth/oauth/wechat     # 微信登录
POST /api/v1/auth/bindoauth        # 绑定第三方账号（需登录）
```

**流程**：

```
1. 前端跳转第三方授权页
2. 用户授权后回调，带 code
3. 后端用 code 换取 access_token
4. 用 access_token 获取用户信息
5. 查询绑定关系：
   - 已绑定 → 直接登录，返回 Token
   - 未绑定 → 自动注册或引导绑定现有账号
```

### Q16: 如何实现多设备登录互踢？

**答：**

**方案一：Token 黑名单**

```go
// 登录时将旧 Token 加入黑名单
func Login(userId int64) {
    // 1. 获取用户当前的 Token
    oldToken := redis.Get("user:token:" + strconv.FormatInt(userId, 10))
    
    // 2. 将旧 Token 加入黑名单
    if oldToken != "" {
        redis.SetEx("blacklist:" + oldToken, "1", tokenExpireTime)
    }
    
    // 3. 生成新 Token 并记录
    newToken := generateToken(userId)
    redis.Set("user:token:" + strconv.FormatInt(userId, 10), newToken)
}

// 验证时检查黑名单
func ValidateToken(token string) bool {
    if redis.Exists("blacklist:" + token) {
        return false  // Token 已被踢出
    }
    return parseAndValidate(token)
}
```

**方案二：Token 版本号**

```go
// 用户表增加 token_version 字段
type User struct {
    // ...
    TokenVersion int64 `db:"token_version"`
}

// Token 中包含版本号
type Claims struct {
    UserId       int64
    TokenVersion int64  // Token 版本
}

// 登录时更新版本号
func Login(userId int64) {
    // UPDATE user SET token_version = token_version + 1 WHERE id = ?
}

// 验证时比对版本号
func ValidateToken(claims Claims) bool {
    user := findUser(claims.UserId)
    return claims.TokenVersion == user.TokenVersion
}
```

**方案二更优**：不需要维护黑名单，数据库一个字段搞定。

---

## 七、总结

面试中关于 Auth 服务，重点考察：

1. **架构设计**：分层、依赖注入、代码组织
2. **安全机制**：JWT、bcrypt、验证码防护
3. **缓存策略**：Cache-Aside、一致性保证
4. **性能优化**：索引、缓存、异步
5. **问题排查**：日志分析、Redis 检查
6. **扩展能力**：第三方登录、多设备管理

记住：**不仅要知道怎么做，还要知道为什么这么做**。

