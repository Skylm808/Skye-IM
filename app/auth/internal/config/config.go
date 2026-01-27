// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package config

import (
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	// User RPC 配置
	UserRpc zrpc.RpcClientConf

	// MySQL 配置（临时保留，后续可删除）
	MySQL struct {
		DataSource string
	}

	// Redis 缓存配置（用于Model缓存，临时保留）
	Cache cache.CacheConf

	// Redis 配置（用于验证码等）
	Redis redis.RedisConf

	// JWT 配置
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}

	// RefreshToken 配置
	RefreshToken struct {
		Secret string
		Expire int64
	}

	// 邮箱配置
	Email struct {
		Host     string // SMTP服务器地址
		Port     int    // SMTP端口
		Username string // 邮箱账号
		Password string // 邮箱授权码
		From     string // 发件人名称
	}

	Captcha struct {
		Expire int64 // 验证码过期时间（秒）
		Length int   // 验证码长度
	}
	Etcd discov.EtcdConf
}
