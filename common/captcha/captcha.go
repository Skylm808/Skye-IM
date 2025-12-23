package captcha

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// CaptchaType 验证码用途类型
type CaptchaType string

const (
	// CaptchaTypeRegister 注册验证码
	CaptchaTypeRegister CaptchaType = "register"
	// CaptchaTypeReset 重置密码验证码
	CaptchaTypeReset CaptchaType = "reset"
)

// 验证码Redis Key前缀
func captchaKey(captchaType CaptchaType, email string) string {
	return fmt.Sprintf("captcha:%s:email:%s", captchaType, email)
}

// 发送频率限制Key前缀
func sendLimitKey(captchaType CaptchaType, email string) string {
	return fmt.Sprintf("captcha:%s:limit:%s", captchaType, email)
}

// Service 验证码服务
type Service struct {
	redis  *redis.Redis
	expire int // 验证码过期时间（秒）
	length int // 验证码长度
}

// NewService 创建验证码服务
func NewService(redisConf redis.RedisConf, expire int, length int) *Service {
	if expire <= 0 {
		expire = 300 // 默认5分钟
	}
	if length <= 0 {
		length = 6 // 默认6位
	}
	return &Service{
		redis:  redis.MustNewRedis(redisConf),
		expire: expire,
		length: length,
	}
}

// Generate 生成验证码
func (s *Service) Generate() string {
	rand.Seed(time.Now().UnixNano())
	code := ""
	for i := 0; i < s.length; i++ {
		code += fmt.Sprintf("%d", rand.Intn(10))
	}
	return code
}

// Store 存储验证码（带用途类型）
func (s *Service) Store(ctx context.Context, captchaType CaptchaType, email, code string) error {
	key := captchaKey(captchaType, email)
	return s.redis.SetexCtx(ctx, key, code, s.expire)
}

// Verify 验证验证码（带用途类型）
func (s *Service) Verify(ctx context.Context, captchaType CaptchaType, email, code string) (bool, error) {
	key := captchaKey(captchaType, email)
	storedCode, err := s.redis.GetCtx(ctx, key)
	if err != nil {
		return false, err
	}
	if storedCode == "" {
		return false, nil // 验证码不存在或已过期
	}
	if storedCode != code {
		return false, nil // 验证码不匹配
	}
	// 验证成功后删除验证码
	_, _ = s.redis.DelCtx(ctx, key)
	return true, nil
}

// Delete 删除验证码
func (s *Service) Delete(ctx context.Context, captchaType CaptchaType, email string) error {
	key := captchaKey(captchaType, email)
	_, err := s.redis.DelCtx(ctx, key)
	return err
}

// CheckSendLimit 检查发送频率限制（60秒内只能发送一次）
func (s *Service) CheckSendLimit(ctx context.Context, captchaType CaptchaType, email string) (bool, error) {
	key := sendLimitKey(captchaType, email)
	exists, err := s.redis.ExistsCtx(ctx, key)
	if err != nil {
		return false, err
	}
	return !exists, nil // 不存在表示可以发送
}

// SetSendLimit 设置发送频率限制
func (s *Service) SetSendLimit(ctx context.Context, captchaType CaptchaType, email string) error {
	key := sendLimitKey(captchaType, email)
	return s.redis.SetexCtx(ctx, key, "1", 60) // 60秒限制
}

// GetTTL 获取验证码剩余有效时间
func (s *Service) GetTTL(ctx context.Context, captchaType CaptchaType, email string) (int, error) {
	key := captchaKey(captchaType, email)
	return s.redis.TtlCtx(ctx, key)
}
