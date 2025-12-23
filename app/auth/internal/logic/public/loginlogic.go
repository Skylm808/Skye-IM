// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package public

import (
	"context"
	"database/sql"
	"strings"

	"SkyeIM/common/errorx"
	"SkyeIM/common/jwt"
	"SkyeIM/common/utils"
	"auth/internal/svc"
	"auth/internal/types"
	"auth/model"

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

	// 2. 查找用户（支持用户名、手机号、邮箱登录）
	var user *model.User

	// 先尝试用户名
	user, err = l.svcCtx.UserModel.FindOneByUsername(l.ctx, req.Username)
	if err == model.ErrNotFound {
		// 尝试手机号
		user, err = l.svcCtx.UserModel.FindOneByPhone(l.ctx, sql.NullString{String: req.Username, Valid: true})
	}
	if err == model.ErrNotFound {
		// 尝试邮箱
		user, err = l.svcCtx.UserModel.FindOneByEmail(l.ctx, sql.NullString{String: req.Username, Valid: true})
	}

	if err != nil {
		if err == model.ErrNotFound {
			return nil, errorx.ErrUserNotFound
		}
		l.Logger.Errorf("查询用户失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	// 3. 检查用户状态
	if user.Status == 0 {
		return nil, errorx.ErrUserDisabled
	}

	// 4. 验证密码
	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, errorx.ErrPasswordWrong
	}

	// 5. 生成Token（注意：user.Id 是 uint64，需要转换为 int64）
	tokenPair, err := jwt.GenerateTokenPair(
		int64(user.Id),
		user.Username,
		l.svcCtx.Config.Auth.AccessSecret,
		l.svcCtx.Config.Auth.AccessExpire,
		l.svcCtx.Config.RefreshToken.Secret,
		l.svcCtx.Config.RefreshToken.Expire,
	)
	if err != nil {
		l.Logger.Errorf("生成Token失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	l.Logger.Infof("用户登录成功: userId=%d, username=%s", user.Id, user.Username)

	return &types.TokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}, nil
}
