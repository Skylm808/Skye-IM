package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	MinIO struct {
		Endpoint        string // MinIO内部访问地址（容器间通信）
		PublicEndpoint  string // MinIO公共访问地址（前端访问）
		AccessKeyID     string
		SecretAccessKey string
		UseSSL          bool
		Buckets         struct {
			Avatar string
			Image  string
			File   string
			Video  string
			Voice  string
		}
	}
	Upload struct {
		MaxImageSize      int64
		MaxFileSize       int64
		MaxAvatarSize     int64
		AllowedImageTypes []string
		AllowedFileTypes  []string
	}
}
