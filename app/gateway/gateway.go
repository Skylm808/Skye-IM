package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"
)

// Config 配置结构体
type Config struct {
	Host      string          // Gateway监听地址（0.0.0.0）
	Port      int             // Gateway监听端口（8080）
	Etcd      discov.EtcdConf // etcd配置（服务发现）
	Auth      AuthConfig      // JWT配置
	WhiteList []string        `json:",optional"` // 白名单配置（不需要JWT鉴权的接口）
	Cors      CorsConfig      // CORS跨域配置
	Log       logx.LogConf    // 日志配置
}

// JWT认证配置
type AuthConfig struct {
	AccessSecret string // JWT密钥（Skylm-im-secret-key）
	AccessExpire int64  // AccessToken过期时间（7200秒=2h）
}

// CORS跨域配置
type CorsConfig struct {
	AllowOrigins     []string // 允许的来源
	AllowMethods     []string // 允许的HTTP方法
	AllowHeaders     []string // 允许的请求头
	ExposeHeaders    []string // 暴露的响应头
	AllowCredentials bool     // 是否允许携带凭证
	MaxAge           int      // 预检缓存时间
}

// =============== Gateway核心结构体 ===============
type Gateway struct {
	config      Config           // 配置（完整的Config）
	whiteList   []*regexp.Regexp // 编译后的白名单正则
	subscribers sync.Map         // 路由表：map[serviceName → *discov.Subscriber]，惰性初始化后长驻内存，持续Watch etcd
}

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()

	var config Config
	conf.MustLoad(*configFile, &config)

	// 初始化日志
	logx.MustSetup(config.Log)
	defer logx.Close()

	gateway := NewGateway(config)

	// 添加CORS中间件
	handler := corsMiddleware(config.Cors)(gateway)

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	logx.Infof("Gateway 启动在端口: %d", config.Port)
	logx.Info("支持的服务: auth-api, user-api, friend-api, message-api，group-api")
	log.Fatal(http.ListenAndServe(addr, handler))
}

// NewGateway 创建网关实例
func NewGateway(config Config) *Gateway {
	g := &Gateway{
		config:    config,
		whiteList: make([]*regexp.Regexp, 0),
	}

	// 编译白名单正则
	for _, pattern := range config.WhiteList {
		re, err := regexp.Compile(pattern)
		if err != nil {
			logx.Errorf("白名单正则编译失败: %s, err: %v", pattern, err)
			continue
		}
		g.whiteList = append(g.whiteList, re)
	}

	logx.Infof("已连接etcd: %v", config.Etcd.Hosts)

	return g
}

// ServeHTTP 实现http.Handler接口
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// 1. 鉴权（白名单外的接口需要JWT验证）
	if !g.isWhiteListed(r.URL.Path) {
		if err := g.authJWT(r); err != nil {
			logx.Errorf("鉴权失败: path=%s, err=%v", r.URL.Path, err)
			http.Error(w, fmt.Sprintf("鉴权失败: %v", err), http.StatusUnauthorized)
			return
		}
	}

	// 2. 提取服务名（从URL路径）
	serviceName := g.extractServiceName(r.URL.Path)
	if serviceName == "" {
		logx.Errorf("无法解析服务名: path=%s", r.URL.Path)
		http.Error(w, "无法解析服务名", http.StatusBadRequest)
		return
	}

	// 3. 获取后端服务地址
	serviceAddr, err := g.getServiceAddr(serviceName)
	if err != nil {
		logx.Errorf("服务发现失败: service=%s, err=%v", serviceName, err)
		http.Error(w, fmt.Sprintf("服务不可用: %v", err), http.StatusServiceUnavailable)
		return
	}

	// 4. 执行反向代理
	g.proxyRequest(w, r, serviceAddr, serviceName)

	// 记录请求信息
	duration := time.Since(startTime)
	logx.Infof("请求完成: %s %s → %s, 耗时: %v", r.Method, r.URL.Path, serviceName, duration)
}

// isWhiteListed 检查URL是否在白名单
func (g *Gateway) isWhiteListed(path string) bool {
	for _, re := range g.whiteList {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}

// authJWT JWT认证
func (g *Gateway) authJWT(r *http.Request) error {
	// 从Header获取Token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return fmt.Errorf("缺少Authorization header")
	}

	// 提取Bearer Token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return fmt.Errorf("Authorization格式错误，应为: Bearer {token}")
	}

	tokenString := parts[1]
	if tokenString == "" {
		return fmt.Errorf("Token为空")
	}

	// 解析JWT Token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(g.config.Auth.AccessSecret), nil
	})

	if err != nil {
		return fmt.Errorf("Token解析失败: %v", err)
	}

	if !token.Valid {
		return fmt.Errorf("Token无效")
	}

	// 提取claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// 可选：注入用户信息到Header（供后端服务使用）
		if userId, ok := claims["userId"]; ok {
			r.Header.Set("X-User-Id", fmt.Sprintf("%v", userId))
		}
		if username, ok := claims["username"]; ok {
			r.Header.Set("X-Username", fmt.Sprintf("%v", username))
		}
	}

	return nil
}

// extractServiceName 从URL提取服务名
// /api/v1/user/profile → "user-api"
// /api/v1/auth/login → "auth-api"
func (g *Gateway) extractServiceName(path string) string {
	// 正则: /api/v1/(service)/...
	re := regexp.MustCompile(`^/api/v1/([^/]+)`)
	matches := re.FindStringSubmatch(path)

	if len(matches) >= 2 {
		return matches[1] + "-api" // user → user-api
	}

	return ""
}

// getOrCreateSubscriber 惰性初始化并缓存 Subscriber。
// 首次调用时创建与 etcd 的长连接并持续 Watch 服务变更；
// 后续调用直接复用已有 Subscriber，避免重复建连。
func (g *Gateway) getOrCreateSubscriber(serviceName string) (*discov.Subscriber, error) {
	// 快路径：已存在直接返回
	if v, ok := g.subscribers.Load(serviceName); ok {
		return v.(*discov.Subscriber), nil
	}

	// 慢路径：首次创建 Subscriber（建立 etcd Watch 长连接）
	sub, err := discov.NewSubscriber(g.config.Etcd.Hosts, serviceName)
	if err != nil {
		return nil, fmt.Errorf("订阅服务[%s]失败: %v", serviceName, err)
	}

	// LoadOrStore 保证并发首次请求只有一个 Subscriber 存入，防止重复建连
	actual, _ := g.subscribers.LoadOrStore(serviceName, sub)
	logx.Infof("[Gateway] 已建立服务[%s]的 etcd Watch 长连接", serviceName)
	return actual.(*discov.Subscriber), nil
}

// getServiceAddr 从路由表（etcd Watch 缓存）获取服务地址，随机负载均衡
func (g *Gateway) getServiceAddr(serviceName string) (string, error) {
	sub, err := g.getOrCreateSubscriber(serviceName)
	if err != nil {
		return "", err
	}

	// Values() 直接返回 Subscriber 内部由 etcd Watch 实时维护的实例列表
	values := sub.Values()
	if len(values) == 0 {
		return "", fmt.Errorf("服务 %s 无可用实例", serviceName)
	}

	// 随机负载均衡
	serviceAddr := values[rand.Intn(len(values))]
	logx.Infof("[Gateway] 路由: %s -> %s（共%d个实例）", serviceName, serviceAddr, len(values))
	return serviceAddr, nil
}

// proxyRequest 执行反向代理
func (g *Gateway) proxyRequest(w http.ResponseWriter, r *http.Request, targetAddr, serviceName string) {
	// 解析目标地址
	target, err := url.Parse("http://" + targetAddr)
	if err != nil {
		logx.Errorf("目标地址解析失败: addr=%s, err=%v", targetAddr, err)
		http.Error(w, "目标地址解析失败", http.StatusInternalServerError)
		return
	}

	// 创建反向代理
	proxy := httputil.NewSingleHostReverseProxy(target)

	// 自定义Director：修改请求
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req) // 先执行原始逻辑

		// 重写Host
		req.Host = target.Host
		req.URL.Host = target.Host
		req.URL.Scheme = target.Scheme

		// 设置转发信息
		req.Header.Set("X-Forwarded-Host", r.Host)      // 原始Host
		req.Header.Set("X-Origin-Host", target.Host)    // 目标Host
		req.Header.Set("X-Forwarded-For", r.RemoteAddr) // 客户端IP
	}

	// 自定义错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logx.Errorf("代理错误: service=%s, addr=%s, err=%v", serviceName, targetAddr, err)
		http.Error(w, fmt.Sprintf("后端服务不可用: %v", err), http.StatusBadGateway)
	}

	// 执行代理
	logx.Infof("转发请求: %s %s → %s (%s)", r.Method, r.URL.Path, serviceName, targetAddr)
	proxy.ServeHTTP(w, r)
	// 自动转发请求并返回响应
}

// corsMiddleware CORS中间件
func corsMiddleware(config CorsConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// 检查origin是否在白名单
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

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))

			if len(config.ExposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
			}

			// 处理OPTIONS预检请求
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
