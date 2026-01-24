// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package public

import (
	"context"
	"strings"

	"SkyeIM/app/user/rpc/userClient"
	"SkyeIM/common/errorx"
	"SkyeIM/common/jwt"
	"auth/internal/svc"
	"auth/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 用户登录
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginRequest) (resp *types.TokenResponse, err error) {
	// 1. 参数校验
	if err := l.svcCtx.Validator.StructCtx(l.ctx, req); err != nil {
		return nil, errorx.NewCodeError(errorx.CodeParam, err.Error())
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	// 2. 通过User RPC验证密码
	verifyResp, err := l.svcCtx.UserRpc.VerifyPassword(l.ctx, &userClient.VerifyPasswordRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		l.Logger.Errorf("RPC验证密码失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	if !verifyResp.Success {
		switch verifyResp.Message {
		case "用户不存在":
			return nil, errorx.ErrUserNotFound
		case "用户已被禁用":
			return nil, errorx.ErrUserDisabled
		case "密码错误":
			return nil, errorx.ErrPasswordWrong
		default:
			return nil, errorx.NewCodeError(errorx.CodeUnknown, verifyResp.Message)
		}
	}

	// 3. 生成Token
	tokenPair, err := jwt.GenerateTokenPair(
		verifyResp.User.Id,
		verifyResp.User.Username,
		l.svcCtx.Config.Auth.AccessSecret,
		l.svcCtx.Config.Auth.AccessExpire,
		l.svcCtx.Config.RefreshToken.Secret,
		l.svcCtx.Config.RefreshToken.Expire,
	)
	if err != nil {
		l.Logger.Errorf("生成Token失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	l.Logger.Infof("用户登录成功: userId=%d, username=%s", verifyResp.User.Id, verifyResp.User.Username)

	return &types.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}
