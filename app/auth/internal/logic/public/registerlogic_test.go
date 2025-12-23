package public

import (
	"context"
	"testing"

	"auth/internal/config"
	"auth/internal/svc"
	"auth/internal/types"
)

func TestRegisterLogic_Register(t *testing.T) {
	// 初始化配置（测试用）
	c := config.Config{}
	c.MySQL.DataSource = "root:630630@tcp(127.0.0.1:3306)/im_auth_test?charset=utf8mb4&parseTime=True&loc=Local"
	c.Auth.AccessSecret = "test-secret"
	c.Auth.AccessExpire = 3600
	c.RefreshToken.Secret = "test-refresh-secret"
	c.RefreshToken.Expire = 86400

	ctx := svc.NewServiceContext(c)

	tests := []struct {
		name    string
		req     *types.RegisterRequest
		wantErr bool
	}{
		{
			name: "正常注册",
			req: &types.RegisterRequest{
				Username: "testuser",
				Password: "123456",
				Nickname: "测试用户",
			},
			wantErr: false,
		},
		{
			name: "用户名过短",
			req: &types.RegisterRequest{
				Username: "ab",
				Password: "123456",
			},
			wantErr: true,
		},
		{
			name: "密码过短",
			req: &types.RegisterRequest{
				Username: "testuser2",
				Password: "123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewRegisterLogic(context.Background(), ctx)
			_, err := l.Register(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
