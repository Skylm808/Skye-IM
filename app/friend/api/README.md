# Friend API 服务

好友关系管理的 HTTP API 网关服务，调用 Friend RPC 处理业务逻辑。

## 功能

### 好友申请
| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 发送申请 | POST | `/api/v1/friend/request` | 发送好友申请 |
| 收到的申请 | GET | `/api/v1/friend/request/received` | 获取收到的申请列表 |
| 发出的申请 | GET | `/api/v1/friend/request/sent` | 获取发出的申请列表 |
| 处理申请 | PUT | `/api/v1/friend/request/:requestId` | 同意/拒绝申请 |

### 好友关系
| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 好友列表 | GET | `/api/v1/friend/list` | 获取好友列表 |
| 删除好友 | DELETE | `/api/v1/friend/:friendId` | 删除好友 |
| 更新备注 | PUT | `/api/v1/friend/:friendId/remark` | 更新好友备注 |
| 检查好友 | GET | `/api/v1/friend/:friendId/check` | 检查是否为好友 |

### 黑名单
| 接口 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 设置黑名单 | POST | `/api/v1/friend/blacklist` | 拉黑/取消拉黑 |
| 黑名单列表 | GET | `/api/v1/friend/blacklist` | 获取黑名单列表 |

> 所有接口需要 JWT 认证

## 配置

```yaml
# etc/friend-api.yaml
Name: friend-api
Host: 0.0.0.0
Port: 10200

Auth:
  AccessSecret: "your-jwt-secret"
  AccessExpire: 604800

FriendRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: friend.rpc
```

## 启动

```bash
# 先启动 Friend RPC
cd app/friend/rpc
go run friend.go -f etc/friend.yaml

# 再启动 Friend API
cd app/friend/api
go run friend.go -f etc/friend-api.yaml
```

## 架构

```
前端 (HTTP) → Friend API (10200) → Friend RPC (9200) → MySQL
```

## 目录结构

```
friend/api/
├── etc/                 # 配置文件
├── internal/
│   ├── config/          # 配置结构
│   ├── handler/         # HTTP处理器
│   │   ├── request/     # 好友申请相关
│   │   ├── friend/      # 好友关系相关
│   │   └── blacklist/   # 黑名单相关
│   ├── logic/           # 业务逻辑（调用RPC）
│   ├── svc/             # 服务上下文（注入FriendRpc）
│   └── types/           # 请求/响应类型
├── friend.api           # API定义文件
└── friend.go            # 入口
```
