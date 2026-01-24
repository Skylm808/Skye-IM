// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package public

import (
	"context"
	"strings"

	"SkyeIM/app/user/rpc/userClient"
	"SkyeIM/common/captcha"
	"SkyeIM/common/errorx"
	"SkyeIM/common/utils"
	"auth/internal/svc"
	"auth/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ForgotPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 忘记密码
func NewForgotPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ForgotPasswordLogic {
	return &ForgotPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ForgotPasswordLogic) ForgotPassword(req *types.ForgotPasswordRequest) (resp *types.ForgotPasswordResponse, err error) {
	// 1. 参数校验
	if err := l.svcCtx.Validator.StructCtx(l.ctx, req); err != nil {
		return nil, errorx.NewCodeError(errorx.CodeParam, err.Error())
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	req.Captcha = strings.TrimSpace(req.Captcha)
	req.NewPassword = strings.TrimSpace(req.NewPassword)

	// 2. 验证邮箱验证码（使用 reset 类型）
	valid, err := l.svcCtx.CaptchaService.Verify(l.ctx, captcha.CaptchaTypeReset, email, req.Captcha)
	if err != nil {
		l.Logger.Errorf("验证码校验失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}
	if !valid {
		return nil, errorx.NewCodeError(errorx.CodeParam, "验证码错误或已过期")
	}

	// 3. 通过RPC查找用户
	userResp, err := l.svcCtx.UserRpc.FindUserByField(l.ctx, &userClient.FindUserByFieldRequest{
		FieldType:  "email",
		FieldValue: email,
	})
	if err != nil {
		l.Logger.Errorf("RPC查询用户失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}
	if !userResp.Found {
		return nil, errorx.NewCodeError(errorx.CodeParam, "该邮箱未注册")
	}

	// 4. 检查用户状态
	if userResp.User.Status == 0 {
		return nil, errorx.ErrUserDisabled
	}

	// 5. 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		l.Logger.Errorf("密码加密失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	// 6. 通过RPC更新密码
	updateResp, err := l.svcCtx.UserRpc.UpdatePassword(l.ctx, &userClient.UpdatePasswordRequest{
		UserId:      userResp.User.Id,
		NewPassword: hashedPassword,
	})
	if err != nil || !updateResp.Success {
		l.Logger.Errorf("RPC更新密码失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "重置密码失败")
	}

	l.Logger.Infof("密码重置成功: userId=%d, email=%s", userResp.User.Id, email)

	return &types.ForgotPasswordResponse{
		Message: "密码重置成功，请使用新密码登录",
	}, nil
}
