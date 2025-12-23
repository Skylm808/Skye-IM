package email

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

// Config 邮箱配置
type Config struct {
	Host     string // SMTP服务器地址
	Port     int    // SMTP端口
	Username string // 邮箱账号
	Password string // 邮箱授权码
	From     string // 发件人名称
}

// Sender 邮件发送器
type Sender struct {
	config Config
	dialer *gomail.Dialer
}

// NewSender 创建邮件发送器
func NewSender(config Config) *Sender {
	// 创建邮件拨号器
	dialer := gomail.NewDialer(config.Host, config.Port, config.Username, config.Password)

	return &Sender{
		config: config,
		dialer: dialer,
	}
}

// SendCode 发送验证码邮件
func (s *Sender) SendCode(to string, code string) error {
	subject := "【Skye-IM】邮箱验证码"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .code { font-size: 32px; font-weight: bold; color: #667eea; letter-spacing: 5px; text-align: center; padding: 20px; background: white; border-radius: 8px; margin: 20px 0; }
        .footer { text-align: center; color: #999; font-size: 12px; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Skye-IM 即时通讯</h1>
        </div>
        <div class="content">
            <p>您好！</p>
            <p>您正在进行邮箱验证，验证码为：</p>
            <div class="code">%s</div>
            <p>验证码有效期为 <strong>5分钟</strong>，请勿泄露给他人。</p>
            <p>如果这不是您本人的操作，请忽略此邮件。</p>
        </div>
        <div class="footer">
            <p>此邮件由系统自动发送，请勿回复</p>
        </div>
    </div>
</body>
</html>
`, code)

	return s.SendHTML(to, subject, body)
}

// SendHTML 发送HTML邮件
func (s *Sender) SendHTML(to, subject, body string) error {
	m := gomail.NewMessage()

	// 设置发件人
	m.SetHeader("From", m.FormatAddress(s.config.Username, s.config.From))
	// 设置收件人
	m.SetHeader("To", to)
	// 设置主题
	m.SetHeader("Subject", subject)
	// 设置HTML正文
	m.SetBody("text/html", body)

	// 发送邮件
	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}
