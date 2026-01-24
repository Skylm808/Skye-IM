// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package public

import (
	"context"

	"SkyeIM/app/user/rpc/userClient"
	"SkyeIM/common/errorx"
	"SkyeIM/common/jwt"
	"auth/internal/svc"
	"auth/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 刷新Token
func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(req *types.RefreshTokenRequest) (resp *types.TokenResponse, err error) {
	// 1. 参数校验
	if err := l.svcCtx.Validator.StructCtx(l.ctx, req); err != nil {
		return nil, errorx.NewCodeError(errorx.CodeParam, err.Error())
	}

	// 2. 解析RefreshToken
	claims, err := jwt.ParseToken(req.RefreshToken, l.svcCtx.Config.RefreshToken.Secret)
	if err != nil {
		l.Logger.Errorf("解析RefreshToken失败: %v", err)
		return nil, errorx.ErrRefreshTokenInvalid
	}

	// 3. 验证Token类型
	if !jwt.ValidateTokenType(claims, jwt.RefreshToken) {
		return nil, errorx.ErrRefreshTokenInvalid
	}

	// 4. 验证用户是否存在且状态正常（通过User RPC）
	userResp, err := l.svcCtx.UserRpc.GetUser(l.ctx, &userClient.GetUserRequest{
		Id: claims.UserId,
	})
	if err != nil {
		l.Logger.Errorf("RPC查询用户失败: %v", err)
		return nil, errorx.ErrUserNotFound
	}
	if userResp.User == nil {
		return nil, errorx.ErrUserNotFound
	}
	if userResp.User.Status == 0 {
		return nil, errorx.ErrUserDisabled
	}

	// 5. 生成新的Token对
	tokenPair, err := jwt.GenerateTokenPair(
		userResp.User.Id,
		userResp.User.Username,
		l.svcCtx.Config.Auth.AccessSecret,
		l.svcCtx.Config.Auth.AccessExpire,
		l.svcCtx.Config.RefreshToken.Secret,
		l.svcCtx.Config.RefreshToken.Expire,
	)
	if err != nil {
		l.Logger.Errorf("生成Token失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	l.Logger.Infof("Token刷新成功: userId=%d", userResp.User.Id)

	return &types.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}
