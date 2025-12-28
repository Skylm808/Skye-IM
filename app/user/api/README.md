# User API 服务

用户资料管理的 HTTP API 网关服务。

## 功能

| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 获取个人资料 | GET | `/api/v1/user/profile` | 获取当前登录用户信息 |
| 更新资料 | PUT | `/api/v1/user/profile` | 更新昵称、头像、手机号 |
| 更新头像 | PUT | `/api/v1/user/avatar` | 单独更新头像 |
| 搜索用户 | GET | `/api/v1/user/search` | 按关键词搜索用户 |
| 获取用户 | GET | `/api/v1/user/:id` | 获取指定用户信息 |

> 所有接口需要 JWT 认证

## 配置

```yaml
# etc/user-api.yaml
Name: user-api
Host: 0.0.0.0
Port: 10100

MySQL:
  DataSource: root:password@tcp(127.0.0.1:3306)/im_auth?charset=utf8mb4&parseTime=True&loc=Local

Cache:
  - Host: 127.0.0.1:6379

Auth:
  AccessSecret: "your-jwt-secret"
  AccessExpire: 604800
```

## 启动

```bash
cd app/user/api
go run user.go -f etc/user-api.yaml
```

## 目录结构

```
user/api/
├── etc/                 # 配置文件
├── internal/
│   ├── config/          # 配置结构
│   ├── handler/         # HTTP处理器
│   ├── logic/           # 业务逻辑
│   ├── svc/             # 服务上下文
│   └── types/           # 请求/响应类型
├── user.api             # API定义文件
└── user.go              # 入口
```
