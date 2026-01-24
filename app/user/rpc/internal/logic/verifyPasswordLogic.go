package logic

import (
	"context"
	"database/sql"

	"SkyeIM/app/user/rpc/internal/svc"
	"SkyeIM/app/user/rpc/user"
	"SkyeIM/common/utils"
	"auth/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyPasswordLogic {
	return &VerifyPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 验证用户密码（登录用）
func (l *VerifyPasswordLogic) VerifyPassword(in *user.VerifyPasswordRequest) (*user.VerifyPasswordResponse, error) {
	var foundUser *model.User
	var err error

	// 先尝试用户名
	foundUser, err = l.svcCtx.UserModel.FindOneByUsername(l.ctx, in.Username)
	if err == model.ErrNotFound {
		// 尝试手机号
		foundUser, err = l.svcCtx.UserModel.FindOneByPhone(l.ctx, sql.NullString{String: in.Username, Valid: true})
	}
	if err == model.ErrNotFound {
		// 尝试邮箱
		foundUser, err = l.svcCtx.UserModel.FindOneByEmail(l.ctx, sql.NullString{String: in.Username, Valid: true})
	}

	if err != nil {
		if err == model.ErrNotFound {
			return &user.VerifyPasswordResponse{
				Success: false,
				Message: "用户不存在",
			}, nil
		}
		l.Logger.Errorf("查询用户失败: %v", err)
		return &user.VerifyPasswordResponse{
			Success: false,
			Message: "系统错误",
		}, nil
	}

	// 检查用户状态
	if foundUser.Status == 0 {
		return &user.VerifyPasswordResponse{
			Success: false,
			Message: "用户已被禁用",
		}, nil
	}

	// 验证密码
	if !utils.CheckPassword(in.Password, foundUser.Password) {
		return &user.VerifyPasswordResponse{
			Success: false,
			Message: "密码错误",
		}, nil
	}

	// 验证成功
	return &user.VerifyPasswordResponse{
		Success: true,
		User:    convertToUserInfo(foundUser),
		Message: "验证成功",
	}, nil
}
