package public

import (
	"context"
	"testing"

	"auth/internal/config"
	"auth/internal/svc"
	"auth/internal/types"
)

func TestLoginLogic_Login(t *testing.T) {
	// 初始化配置（测试用）
	c := config.Config{}
	c.MySQL.DataSource = "root:123456@tcp(127.0.0.1:3306)/im_auth_test?charset=utf8mb4&parseTime=True&loc=Local"
	c.Auth.AccessSecret = "test-secret"
	c.Auth.AccessExpire = 3600
	c.RefreshToken.Secret = "test-refresh-secret"
	c.RefreshToken.Expire = 86400

	ctx := svc.NewServiceContext(c)

	tests := []struct {
		name    string
		req     *types.LoginRequest
		wantErr bool
	}{
		{
			name: "用户名为空",
			req: &types.LoginRequest{
				Username: "",
				Password: "123456",
			},
			wantErr: true,
		},
		{
			name: "密码为空",
			req: &types.LoginRequest{
				Username: "testuser",
				Password: "",
			},
			wantErr: true,
		},
		{
			name: "用户不存在",
			req: &types.LoginRequest{
				Username: "nonexistent",
				Password: "123456",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLoginLogic(context.Background(), ctx)
			_, err := l.Login(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

