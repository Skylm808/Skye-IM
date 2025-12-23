// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package public

import (
	"context"
	"database/sql"
	"strings"

	"SkyeIM/common/captcha"
	"SkyeIM/common/errorx"
	"SkyeIM/common/jwt"
	"SkyeIM/common/utils"
	"auth/internal/svc"
	"auth/internal/types"
	"auth/model"

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
	_, err = l.svcCtx.UserModel.FindOneByUsername(l.ctx, req.Username)
	if err == nil {
		return nil, errorx.ErrUsernameExists
	}
	if err != model.ErrNotFound {
		l.Logger.Errorf("查询用户失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	// 4. 检查邮箱是否已存在
	_, err = l.svcCtx.UserModel.FindOneByEmail(l.ctx, sql.NullString{String: req.Email, Valid: true})
	if err == nil {
		return nil, errorx.ErrEmailExists
	}
	if err != model.ErrNotFound {
		l.Logger.Errorf("查询邮箱失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	// 5. 检查手机号是否已存在（如果提供了手机号）
	if req.Phone != "" {
		_, err = l.svcCtx.UserModel.FindOneByPhone(l.ctx, sql.NullString{String: req.Phone, Valid: true})
		if err == nil {
			return nil, errorx.ErrPhoneExists
		}
		if err != model.ErrNotFound {
			l.Logger.Errorf("查询手机号失败: %v", err)
			return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
		}
	}

	// 6. 密码加密
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		l.Logger.Errorf("密码加密失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	// 7. 设置昵称
	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}

	// 8. 创建用户
	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Email:    sql.NullString{String: req.Email, Valid: true},
		Phone:    sql.NullString{String: req.Phone, Valid: req.Phone != ""},
		Nickname: nickname,
		Avatar:   "",
		Status:   1,
	}

	result, err := l.svcCtx.UserModel.Insert(l.ctx, user)
	if err != nil {
		l.Logger.Errorf("创建用户失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "注册失败")
	}

	// 9. 获取新用户ID
	userId, err := result.LastInsertId()
	if err != nil {
		l.Logger.Errorf("获取用户ID失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	// 10. 生成Token
	tokenPair, err := jwt.GenerateTokenPair(
		userId,
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

	l.Logger.Infof("用户注册成功: userId=%d, username=%s, email=%s", userId, req.Username, req.Email)

	return &types.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}
