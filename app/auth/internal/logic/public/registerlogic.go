// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package public

import (
	"context"
	"strings"

	"SkyeIM/app/user/rpc/userClient"
	"SkyeIM/common/captcha"
	"SkyeIM/common/errorx"
	"SkyeIM/common/jwt"
	"SkyeIM/common/utils"
	"auth/internal/svc"
	"auth/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户注册
func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.TokenResponse, err error) {
	// 1. 参数校验
	if err := l.svcCtx.Validator.StructCtx(l.ctx, req); err != nil {
		return nil, errorx.NewCodeError(errorx.CodeParam, err.Error())
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Captcha = strings.TrimSpace(req.Captcha)

	// 2. 验证邮箱验证码（使用 register 类型）
	l.Logger.Infof("验证验证码: email=%s, captcha=%s", req.Email, req.Captcha)
	valid, err := l.svcCtx.CaptchaService.Verify(l.ctx, captcha.CaptchaTypeRegister, req.Email, req.Captcha)
	if err != nil {
		l.Logger.Errorf("验证码校验失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}
	if !valid {
		l.Logger.Errorf("验证码验证失败: email=%s, captcha=%s", req.Email, req.Captcha)
		return nil, errorx.NewCodeError(errorx.CodeParam, "验证码错误或已过期")
	}
	l.Logger.Infof("验证码验证成功: email=%s", req.Email)

	// 3. 检查用户名是否已存在
	userResp, err := l.svcCtx.UserRpc.FindUserByField(l.ctx, &userClient.FindUserByFieldRequest{
		FieldType:  "username",
		FieldValue: req.Username,
	})
	if err != nil {
		l.Logger.Errorf("RPC查询用户名失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}
	if userResp.Found {
		return nil, errorx.ErrUsernameExists
	}

	// 4. 检查邮箱是否已存在
	emailResp, err := l.svcCtx.UserRpc.FindUserByField(l.ctx, &userClient.FindUserByFieldRequest{
		FieldType:  "email",
		FieldValue: req.Email,
	})
	if err != nil {
		l.Logger.Errorf("RPC查询邮箱失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}
	if emailResp.Found {
		return nil, errorx.ErrEmailExists
	}

	// 5. 检查手机号是否已存在（如果提供了手机号）
	if req.Phone != "" {
		phoneResp, err := l.svcCtx.UserRpc.FindUserByField(l.ctx, &userClient.FindUserByFieldRequest{
			FieldType:  "phone",
			FieldValue: req.Phone,
		})
		if err != nil {
			l.Logger.Errorf("RPC查询手机号失败: %v", err)
			return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
		}
		if phoneResp.Found {
			return nil, errorx.ErrPhoneExists
		}
	}

	// 6. 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		l.Logger.Errorf("密码加密失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	// 7. 通过User RPC创建用户
	createResp, err := l.svcCtx.UserRpc.CreateUser(l.ctx, &userClient.CreateUserRequest{
		Username: req.Username,
		Password: hashedPassword,
		Email:    req.Email,
		Phone:    req.Phone,
		Nickname: req.Nickname,
	})
	if err != nil {
		l.Logger.Errorf("RPC创建用户失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "注册失败")
	}

	// 8. 生成Token
	tokenPair, err := jwt.GenerateTokenPair(
		createResp.UserId,
		req.Username,
		l.svcCtx.Config.Auth.AccessSecret,
		l.svcCtx.Config.Auth.AccessExpire,
		l.svcCtx.Config.RefreshToken.Secret,
		l.svcCtx.Config.RefreshToken.Expire,
	)
	if err != nil {
		l.Logger.Errorf("生成Token失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	l.Logger.Infof("用户注册成功: userId=%d, username=%s, email=%s", createResp.UserId, req.Username, req.Email)

	return &types.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}
