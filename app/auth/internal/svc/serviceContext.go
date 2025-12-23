// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package svc

import (
	"SkyeIM/common/captcha"
	"SkyeIM/common/email"
	"auth/internal/config"
	"auth/model"

	"github.com/go-playground/validator/v10"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config         config.Config
	UserModel      model.UserModel
	Validator      *validator.Validate
	EmailSender    *email.Sender
	CaptchaService *captcha.Service
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.MySQL.DataSource)

	// 创建邮件发送器
	emailSender := email.NewSender(email.Config{
		Host:     c.Email.Host,
		Port:     c.Email.Port,
		Username: c.Email.Username,
		Password: c.Email.Password,
		From:     c.Email.From,
	})

	// 创建验证码服务
	captchaService := captcha.NewService(c.Redis, int(c.Captcha.Expire), c.Captcha.Length)

	return &ServiceContext{
		Config:         c,
		UserModel:      model.NewUserModel(conn, c.Cache),
		Validator:      validator.New(),
		EmailSender:    emailSender,
		CaptchaService: captchaService,
	}
}
