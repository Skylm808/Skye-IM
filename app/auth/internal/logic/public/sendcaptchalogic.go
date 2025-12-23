// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package public

import (
	"context"
	"database/sql"
	"strings"

	"SkyeIM/common/captcha"
	"SkyeIM/common/errorx"
	"auth/internal/svc"
	"auth/internal/types"
	"auth/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendCaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 发送邮箱验证码
func NewSendCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendCaptchaLogic {
	return &SendCaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendCaptchaLogic) SendCaptcha(req *types.SendCaptchaRequest) (resp *types.SendCaptchaResponse, err error) {
	// 1. 参数校验
	if err := l.svcCtx.Validator.StructCtx(l.ctx, req); err != nil {
		return nil, errorx.NewCodeError(errorx.CodeParam, err.Error())
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))
	captchaType := captcha.CaptchaType(req.Type)

	// 2. 根据验证码类型检查邮箱
	_, err = l.svcCtx.UserModel.FindOneByEmail(l.ctx, sql.NullString{String: email, Valid: true})

	if captchaType == captcha.CaptchaTypeRegister {
		// 注册：邮箱不能已存在
		if err == nil {
			return nil, errorx.ErrEmailExists
		}
		if err != model.ErrNotFound {
			l.Logger.Errorf("查询邮箱失败: %v", err)
			return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
		}
	} else if captchaType == captcha.CaptchaTypeReset {
		// 重置密码：邮箱必须存在
		if err == model.ErrNotFound {
			return nil, errorx.NewCodeError(errorx.CodeParam, "该邮箱未注册")
		}
		if err != nil {
			l.Logger.Errorf("查询邮箱失败: %v", err)
			return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
		}
	}

	// 3. 检查发送频率限制（60秒内只能发送一次）
	canSend, err := l.svcCtx.CaptchaService.CheckSendLimit(l.ctx, captchaType, email)
	if err != nil {
		l.Logger.Errorf("检查发送限制失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}
	if !canSend {
		return nil, errorx.NewCodeError(errorx.CodeParam, "验证码发送过于频繁，请60秒后重试")
	}

	// 4. 生成验证码
	code := l.svcCtx.CaptchaService.Generate()

	// 5. 发送邮件
	if err := l.svcCtx.EmailSender.SendCode(email, code); err != nil {
		l.Logger.Errorf("发送验证码邮件失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "发送验证码失败，请稍后重试")
	}

	// 6. 存储验证码（带类型）
	if err := l.svcCtx.CaptchaService.Store(l.ctx, captchaType, email, code); err != nil {
		l.Logger.Errorf("存储验证码失败: %v", err)
		return nil, errorx.NewCodeError(errorx.CodeUnknown, "系统错误")
	}

	// 7. 设置发送频率限制
	if err := l.svcCtx.CaptchaService.SetSendLimit(l.ctx, captchaType, email); err != nil {
		l.Logger.Errorf("设置发送限制失败: %v", err)
		// 不影响主流程，继续执行
	}

	l.Logger.Infof("验证码发送成功: email=%s, type=%s", email, captchaType)

	return &types.SendCaptchaResponse{
		Message: "验证码已发送到您的邮箱，请注意查收",
	}, nil
}
