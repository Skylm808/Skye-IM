# Friend RPC 服务

好友关系管理的 gRPC 微服务，处理所有好友相关的核心业务逻辑。

## 接口

### 好友申请
| 方法 | 说明 |
|------|------|
| `AddFriendRequest` | 发送好友申请 |
| `GetFriendRequestList` | 获取收到的好友申请 |
| `GetSentRequestList` | 获取发出的好友申请 |
| `HandleFriendRequest` | 处理好友申请（同意/拒绝） |

### 好友关系
| 方法 | 说明 |
|------|------|
| `GetFriendList` | 获取好友列表 |
| `DeleteFriend` | 删除好友（双向删除） |
| `UpdateFriendRemark` | 更新好友备注 |
| `IsFriend` | 检查是否为好友 |

### 黑名单
| 方法 | 说明 |
|------|------|
| `SetBlacklist` | 设置黑名单（拉黑/取消） |
| `GetBlacklist` | 获取黑名单列表 |

## 数据表

```sql
-- im_friend: 好友关系表
CREATE TABLE im_friend (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT NOT NULL,
    friend_id BIGINT NOT NULL,
    remark VARCHAR(50),
    status TINYINT DEFAULT 1,  -- 1正常 2拉黑
    created_at DATETIME,
    UNIQUE KEY (user_id, friend_id)
);

-- im_friend_request: 好友申请表
CREATE TABLE im_friend_request (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    from_user_id BIGINT NOT NULL,
    to_user_id BIGINT NOT NULL,
    message VARCHAR(255),
    status TINYINT DEFAULT 0,  -- 0待处理 1同意 2拒绝
    created_at DATETIME
);
```

## 配置

```yaml
# etc/friend.yaml
Name: friend.rpc
ListenOn: 0.0.0.0:9200

MySQL:
  DataSource: root:password@tcp(127.0.0.1:3306)/im_auth?charset=utf8mb4&parseTime=true&loc=Local

Cache:
  - Host: 127.0.0.1:6379

Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: friend.rpc
```

## 启动

```bash
cd app/friend/rpc
go run friend.go -f etc/friend.yaml
```

## 调用示例

```go
// Friend API 调用 Friend RPC
import "SkyeIM/app/friend/rpc/friendclient"

friendRpc := friendclient.NewFriend(zrpc.MustNewClient(c.FriendRpc))
resp, err := friendRpc.GetFriendList(ctx, &friendclient.GetFriendListReq{
    UserId:   1,
    Page:     1,
    PageSize: 20,
})
```

## 目录结构

```
friend/rpc/
├── etc/                 # 配置文件
├── internal/
│   ├── config/          # 配置结构
│   ├── logic/           # 核心业务逻辑
│   ├── server/          # gRPC服务实现
│   └── svc/             # 服务上下文（注入Model）
├── model/               # 数据库Model
├── friend/              # protobuf生成的代码
├── friendclient/        # RPC客户端
├── friend.proto         # 接口定义
└── friend.go            # 入口
```
