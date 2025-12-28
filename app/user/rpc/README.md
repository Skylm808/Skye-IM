# User RPC 服务

用户信息的 gRPC 微服务，供其他服务内部调用。

## 接口

| 方法 | 说明 | 典型调用方 |
|------|------|-----------|
| `GetUser` | 获取单个用户信息 | Friend服务 |
| `BatchGetUsers` | 批量获取用户信息 | Message服务 |
| `SearchUser` | 搜索用户 | Friend服务 |
| `UpdateUser` | 更新用户信息 | User API |
| `CheckUserExist` | 检查用户是否存在 | Gateway |

## 配置

```yaml
# etc/user.yaml
Name: user.rpc
ListenOn: 0.0.0.0:9001

MySQL:
  DataSource: root:password@tcp(127.0.0.1:3306)/im_auth?charset=utf8mb4&parseTime=true&loc=Local

Cache:
  - Host: 127.0.0.1:6379

Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: user.rpc
```

## 启动

```bash
cd app/user/rpc
go run user.go -f etc/user.yaml
```

## 调用示例

```go
// 其他服务调用 User RPC
import "SkyeIM/app/user/rpc/userClient"

userRpc := userClient.NewUser(zrpc.MustNewClient(c.UserRpc))
resp, err := userRpc.GetUser(ctx, &userClient.GetUserRequest{Id: 123})
```

## 目录结构

```
user/rpc/
├── etc/                 # 配置文件
├── internal/
│   ├── config/          # 配置结构
│   ├── logic/           # 业务逻辑
│   ├── server/          # gRPC服务实现
│   └── svc/             # 服务上下文
├── user/                # protobuf生成的代码
├── userClient/          # RPC客户端
├── user.proto           # 接口定义
└── user.go              # 入口
```
