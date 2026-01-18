# Gateway 网关架构与原理讲解

## 一、服务概述

Gateway 是 SkyeIM 的**统一 API 网关**，作为前端与后端微服务之间的唯一入口，提供路由转发、认证鉴权、服务发现等核心功能。

### 为什么需要网关？

**没有网关的问题**：
```
前端 ─┬─> Auth API (10001)
      ├─> User API (10100)
      ├─> Friend API (10200)
      └─> Message API (10400)
```
- ❌ 前端需要配置多个后端地址
- ❌ 跨域问题需要每个服务单独处理
- ❌ 鉴权逻辑在每个服务重复实现
- ❌ 难以统一限流、监控、日志

**使用网关后**：
```
前端 ──> Gateway (8080) ─┬─> Auth API (10001)
                          ├─> User API (10100)
                          ├─> Friend API (10200)
                          └─> Message API (10400)
```
- ✅ 前端只需配置一个地址
- ✅ CORS 统一在网关处理
- ✅ JWT 鉴权统一验证
- ✅ 便于添加限流、缓存、监控

---

## 二、核心功能

### 2.1 功能列表

| 功能 | 说明 | 实现方式 |
|------|------|---------|
| 路由转发 | 根据 URL 自动路由到后端服务 | 正则匹配 + etcd 服务发现 |
| JWT 鉴权 | 验证 Token 有效性 | jwt-go 库 |
| 服务发现 | 动态发现后端服务地址 | go-zero etcd 客户端 |
| 反向代理 | 转发请求并返回响应 | Go 标准库 httputil |
| CORS 处理 | 跨域请求支持 | 自定义中间件 |
| 白名单 | 部分接口跳过鉴权 | 正则表达式匹配 |

---

## 三、架构设计

### 3.1 整体架构

```
┌─────────────────┐
│   前端应用       │
│ (React/Vue)     │
└────────┬────────┘
         │ HTTP Request
         │ Authorization: Bearer {token}
         ▼
┌─────────────────────────────────────────────┐
│         Gateway (8080)                      │
│  ┌──────────────────────────────────────┐  │
│  │  1. CORS 中间件                      │  │
│  │     - 处理 OPTIONS 预检              │  │
│  │     - 设置 CORS 响应头               │  │
│  └──────────┬───────────────────────────┘  │
│             │                                │
│  ┌──────────▼───────────────────────────┐  │
│  │  2. 白名单检查                       │  │
│  │     - 匹配白名单正则                 │  │
│  │     - 跳过 /auth/login 等接口        │  │
│  └──────────┬───────────────────────────┘  │
│             │ 非白名单                      │
│  ┌──────────▼───────────────────────────┐  │
│  │  3. JWT 鉴权                         │  │
│  │     - 提取 Bearer Token              │  │
│  │     - 验证签名和有效期               │  │
│  │     - 注入用户信息到 Header          │  │
│  └──────────┬───────────────────────────┘  │
│             │                                │
│  ┌──────────▼───────────────────────────┐  │
│  │  4. 服务名提取                       │  │
│  │     - 正则匹配: /api/v1/{service}/.. │  │
│  │     - user → user-api                │  │
│  └──────────┬───────────────────────────┘  │
│             │                                │
│  ┌──────────▼───────────────────────────┐  │
│  │  5. 服务发现                         │  │
│  │     - 静态配置优先                   │  │
│  │     - etcd 查询备用                  │  │
│  └──────────┬───────────────────────────┘  │
│             │                                │
│  ┌──────────▼───────────────────────────┐  │
│  │  6. 反向代理                         │  │
│  │     - 修改目标地址                   │  │
│  │     - 转发请求                       │  │
│  │     - 返回响应                       │  │
│  └──────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
         │
         ▼
┌─────────────────┐
│  后端服务        │
│  (10001-10600)  │
└─────────────────┘
```

### 3.2 请求处理流程

```go
// 伪代码
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 1. CORS 处理（中间件）
    if r.Method == "OPTIONS" {
        return // 预检请求直接返回
    }
    
    // 2. 白名单检查
    if !g.isWhiteListed(r.URL.Path) {
        // 3. JWT 鉴权
        if err := g.authJWT(r); err != nil {
            http.Error(w, "鉴权失败", 401)
            return
        }
    }
    
    // 4. 提取服务名
    serviceName := g.extractServiceName(r.URL.Path)
    // /api/v1/user/profile → "user-api"
    
    // 5. 服务发现
    serviceAddr := g.getServiceAddr(serviceName)
    // "user-api" → "127.0.0.1:10100"
    
    // 6. 反向代理
    g.proxyRequest(w, r, serviceAddr, serviceName)
}
```

---

## 四、核心组件详解

### 4.1 路由解析

**功能**：从 URL 提取服务名

**实现**：
```go
func (g *Gateway) extractServiceName(path string) string {
    // 正则: /api/v1/(service)/...
    re := regexp.MustCompile(`^/api/v1/([^/]+)`)
    matches := re.FindStringSubmatch(path)
    
    if len(matches) >= 2 {
        return matches[1] + "-api" // user → user-api
    }
    
    return ""
}
```

**示例**：
```
/api/v1/user/profile     → user-api
/api/v1/auth/login       → auth-api
/api/v1/friend/list      → friend-api
/api/v1/message/history  → message-api
```

**优点**：
- ✅ 自动识别新服务，无需修改网关代码
- ✅ 统一的 URL 规范

---

### 4.2 服务发现

**功能**：根据服务名查找后端地址

**实现**：混合模式（静态 + 动态）

```go
func (g *Gateway) getServiceAddr(serviceName string) (string, error) {
    // 1. 静态配置（硬编码，API 服务）
    staticServices := map[string]string{
        "auth-api":    "127.0.0.1:10001",
        "user-api":    "127.0.0.1:10100",
        "friend-api":  "127.0.0.1:10200",
        "message-api": "127.0.0.1:10400",
        "group-api":   "127.0.0.1:10500",
        "upload-api":  "127.0.0.1:10600",
    }
    
    if addr, ok := staticServices[serviceName]; ok {
        return addr, nil
    }
    
    // 2. etcd 动态查询（RPC 服务或未配置的服务）
    sub, _ := discov.NewSubscriber(g.config.Etcd.Hosts, serviceName)
    values := sub.Values() // []string{"127.0.0.1:9100", ...}
    
    if len(values) == 0 {
        return "", errors.New("服务无可用实例")
    }
    
    return values[0], nil // 简单负载均衡：取第一个
}
```

**为什么混合？**
- **静态配置（API 服务）**：
  - ✅ 响应快（无需查询 etcd）
  - ✅ 端口固定，不变
  
- **etcd 查询（RPC 服务）**：
  - ✅ 支持动态扩缩容
  - ✅ 服务自动注册/注销

---

### 4.3 JWT 鉴权

**功能**：验证 AccessToken 有效性

**流程**：
```go
func (g *Gateway) authJWT(r *http.Request) error {
    // 1. 提取 Token
    authHeader := r.Header.Get("Authorization")
    // "Bearer eyJhbGciOiJIUzI1NiIs..."
    
    parts := strings.SplitN(authHeader, " ", 2)
    tokenString := parts[1]
    
    // 2. 解析 JWT
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(g.config.Auth.AccessSecret), nil
    })
    
    // 3. 验证有效性
    if !token.Valid {
        return errors.New("Token无效")
    }
    
    // 4. 提取 Claims（可选）
    claims := token.Claims.(jwt.MapClaims)
    userId := claims["userId"]
    username := claims["username"]
    
    // 5. 注入用户信息到 Header（供后端使用）
    r.Header.Set("X-User-Id", fmt.Sprintf("%v", userId))
    r.Header.Set("X-Username", fmt.Sprintf("%v", username))
    
    return nil
}
```

**注入的 Header**：
```
X-User-Id: 1001
X-Username: skylm808
```

后端服务可以直接使用，无需再次解析 JWT。

---

### 4.4 反向代理

**功能**：将请求转发到后端服务

**实现**：使用 Go 标准库 `httputil.ReverseProxy`

```go
func (g *Gateway) proxyRequest(w http.ResponseWriter, r *http.Request, 
                                targetAddr, serviceName string) {
    // 1. 解析目标地址
    target, _ := url.Parse("http://" + targetAddr)
    
    // 2. 创建反向代理
    proxy := httputil.NewSingleHostReverseProxy(target)
    
    // 3. 自定义 Director（修改请求）
    proxy.Director = func(req *http.Request) {
        req.Host = target.Host
        req.URL.Host = target.Host
        req.URL.Scheme = "http"
        
        // 设置转发信息
        req.Header.Set("X-Forwarded-Host", r.Host)
        req.Header.Set("X-Forwarded-For", r.RemoteAddr)
    }
    
    // 4. 自定义错误处理
    proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
        http.Error(w, "后端服务不可用", 502)
    }
    
    // 5. 执行代理
    proxy.ServeHTTP(w, r)
}
```

**原理**：
1. 修改请求的目标地址（Host、URL）
2. 转发请求到后端
3. 后端响应自动返回给前端

---

### 4.5 白名单机制

**功能**：部分接口跳过 JWT 鉴权

**实现**：
```go
type Gateway struct {
    whiteList []*regexp.Regexp // 编译后的正则
}

// 初始化时编译正则
func NewGateway(config Config) *Gateway {
    g := &Gateway{whiteList: make([]*regexp.Regexp, 0)}
    
    for _, pattern := range config.WhiteList {
        re, _ := regexp.Compile(pattern)
        g.whiteList = append(g.whiteList, re)
    }
    
    return g
}

// 检查是否在白名单
func (g *Gateway) isWhiteListed(path string) bool {
    for _, re := range g.whiteList {
        if re.MatchString(path) {
            return true
        }
    }
    return false
}
```

**配置示例**：
```yaml
WhiteList:
  - ^/api/v1/auth/login$           # 精确匹配
  - ^/api/v1/auth/register$        # 精确匹配
  - ^/api/v1/auth/captcha/send$    # 精确匹配
  - ^/api/v1/group/public/.*$      # 模糊匹配
```

**正则说明**：
- `^` - 开头
- `$` - 结尾
- `.*` - 任意字符

---

### 4.6 CORS 中间件

**功能**：处理跨域请求

**实现**：
```go
func corsMiddleware(config CorsConfig) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")
            
            // 检查 origin 是否在白名单
            allowed := false
            for _, allowedOrigin := range config.AllowOrigins {
                if origin == allowedOrigin || allowedOrigin == "*" {
                    allowed = true
                    break
                }
            }
            
            if allowed {
                w.Header().Set("Access-Control-Allow-Origin", origin)
            }
            
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            w.Header().Set("Access-Control-Allow-Credentials", "true")
            
            // 处理 OPTIONS 预检请求
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

**CORS 流程**：
1. 浏览器发送 `OPTIONS` 预检请求
2. Gateway 返回允许的 Headers 和 Methods
3. 浏览器发送实际请求
4. Gateway 在响应头设置 `Access-Control-Allow-Origin`

---

## 五、配置说明

### 5.1 完整配置示例

```yaml
# etc/gateway.yaml

Host: 0.0.0.0             # 监听地址
Port: 8080                # 监听端口

# etcd 服务发现配置
Etcd:
  Hosts:
    - 127.0.0.1:2379      # etcd 地址
  Key: ""                 # Gateway 不注册到 etcd

# JWT 配置（必须与后端服务一致）
Auth:
  AccessSecret: "Skylm-im-secret-key"
  AccessExpire: 604800    # 7 天

# 白名单（正则表达式）
WhiteList:
  - ^/api/v1/auth/login$
  - ^/api/v1/auth/register$
  - ^/api/v1/auth/captcha/send$
  - ^/api/v1/auth/password/forgot$
  - ^/api/v1/auth/refresh$

# CORS 配置
Cors:
  AllowOrigins:
    - "http://localhost:3000"    # React 默认端口
    - "http://localhost:5173"    # Vite 默认端口
  AllowMethods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
    - "OPTIONS"
  AllowHeaders:
    - "Content-Type"
    - "Authorization"
    - "X-Requested-With"
  ExposeHeaders:
    - "Content-Length"
  AllowCredentials: true
  MaxAge: 3600              # 预检缓存时间（秒）

# 日志配置
Log:
  ServiceName: gateway
  Mode: console             # console 或 file
  Level: info               # debug, info, error
```

---

## 六、部署与运维

### 6.1 启动顺序

```bash
# 1. 启动 etcd
etcd

# 2. 启动后端服务（任意顺序）
cd app/auth && go run auth.go
cd app/user/api && go run user.go
cd app/friend/api && go run friend.go
cd app/message/api && go run message.go
cd app/group/api && go run group.go
cd app/upload/api && go run upload.go

# 3. 启动 Gateway（最后启动）
cd app/gateway && go run gateway.go
```

### 6.2 监控指标

**关键指标**：
- 请求总量（QPS）
- 平均响应时间
- 鉴权失败次数
- 服务不可用次数
- etcd 查询耗时

**日志示例**：
```
请求完成: POST /api/v1/auth/login → auth-api, 耗时: 45ms
转发请求: GET /api/v1/user/profile → user-api (127.0.0.1:10100)
鉴权失败: path=/api/v1/friend/list, err=Token无效
```

### 6.3 故障处理

| 错误 | 原因 | 解决方案 |
|------|------|---------|
| 服务不可用 | 后端服务未启动 | 启动后端服务 |
| 鉴权失败 | AccessSecret 不一致 | 检查配置文件 |
| 无法连接 etcd | etcd 未启动 | 启动 etcd |

---

## 七、性能优化

### 7.1 当前性能

| 操作 | 耗时 |
|------|------|
| JWT 验证 | < 1 ms |
| etcd 查询（静态配置跳过） | 0 ms |
| 反向代理 | < 1 ms |
| **总计** | **约 1-2 ms** |

### 7.2 优化建议

**1. 服务发现缓存**：
```go
type ServiceCache struct {
    cache map[string]string
    mu    sync.RWMutex
}

func (g *Gateway) getServiceAddr(serviceName string) (string, error) {
    // 先查缓存
    g.cache.mu.RLock()
    if addr, ok := g.cache.cache[serviceName]; ok {
        g.cache.mu.RUnlock()
        return addr, nil
    }
    g.cache.mu.RUnlock()
    
    // 未命中，查询 etcd
    addr := queryEtcd(serviceName)
    
    // 写入缓存
    g.cache.mu.Lock()
    g.cache.cache[serviceName] = addr
    g.cache.mu.Unlock()
    
    return addr, nil
}
```

**2. 连接池复用**：
- 使用 `http.Transport` 连接池
- 避免每次请求创建新连接

**3. 负载均衡**：
```go
// 简单轮询
type RoundRobin struct {
    servers []string
    index   int
    mu      sync.Mutex
}

func (rr *RoundRobin) Next() string {
    rr.mu.Lock()
    defer rr.mu.Unlock()
    
    server := rr.servers[rr.index]
    rr.index = (rr.index + 1) % len(rr.servers)
    return server
}
```

---

## 八、总结

### 优点

- ✅ **统一入口**：前端只需配置一个地址
- ✅ **统一鉴权**：JWT 验证集中处理
- ✅ **服务解耦**：后端服务无需关心 CORS
- ✅ **易于扩展**：新服务自动识别
- ✅ **性能优秀**：延迟 < 2ms

### 缺点

- ❌ **单点故障**：Gateway 挂掉全站不可用（需要高可用）
- ❌ **性能瓶颈**：所有流量经过 Gateway（需要水平扩展）

### 改进方向

- [ ] 实现 Gateway 高可用（多实例 + 负载均衡）
- [ ] 添加限流、熔断功能
- [ ] 实现更智能的负载均衡（权重、健康检查）
- [ ] 添加 API 缓存（GET请求结果缓存）
- [ ] 集成 Prometheus 监控
- [ ] 支持 WebSocket 代理

---

**文档作者**: Skylm  
**最后更新**: 2026-01-13  
**相关文档**: [Gateway API 文档](./GATEWAY_API文档.md)
