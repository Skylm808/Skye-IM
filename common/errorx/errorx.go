package errorx

import "fmt"

// CodeError 业务错误
type CodeError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// 错误码定义
const (
	// 通用错误 10000-10099
	CodeSuccess      = 0
	CodeUnknown      = 10000
	CodeParam        = 10001
	CodeUnauthorized = 10002
	CodeForbidden    = 10003
	CodeNotFound     = 10004

	// 用户相关错误 10100-10199
	CodeUserNotFound         = 10100
	CodeUserExists           = 10101
	CodePasswordWrong        = 10102
	CodeUserDisabled         = 10103
	CodePhoneExists          = 10104
	CodeEmailExists          = 10105
	CodeUsernameExists       = 10106
	CodeTokenInvalid         = 10107
	CodeTokenExpired         = 10108
	CodeRefreshTokenInvalid  = 10109
)

// 预定义错误
var (
	ErrUnknown             = NewCodeError(CodeUnknown, "未知错误")
	ErrParam               = NewCodeError(CodeParam, "参数错误")
	ErrUnauthorized        = NewCodeError(CodeUnauthorized, "未授权")
	ErrForbidden           = NewCodeError(CodeForbidden, "禁止访问")
	ErrNotFound            = NewCodeError(CodeNotFound, "资源不存在")
	ErrUserNotFound        = NewCodeError(CodeUserNotFound, "用户不存在")
	ErrUserExists          = NewCodeError(CodeUserExists, "用户已存在")
	ErrPasswordWrong       = NewCodeError(CodePasswordWrong, "密码错误")
	ErrUserDisabled        = NewCodeError(CodeUserDisabled, "用户已被禁用")
	ErrPhoneExists         = NewCodeError(CodePhoneExists, "手机号已被注册")
	ErrEmailExists         = NewCodeError(CodeEmailExists, "邮箱已被注册")
	ErrUsernameExists      = NewCodeError(CodeUsernameExists, "用户名已被注册")
	ErrTokenInvalid        = NewCodeError(CodeTokenInvalid, "Token无效")
	ErrTokenExpired        = NewCodeError(CodeTokenExpired, "Token已过期")
	ErrRefreshTokenInvalid = NewCodeError(CodeRefreshTokenInvalid, "刷新Token无效")
)

// NewCodeError 创建业务错误
func NewCodeError(code int, msg string) *CodeError {
	return &CodeError{
		Code:    code,
		Message: msg,
	}
}

// Error 实现error接口
func (e *CodeError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// GetCode 获取错误码
func (e *CodeError) GetCode() int {
	return e.Code
}

// GetMessage 获取错误信息
func (e *CodeError) GetMessage() string {
	return e.Message
}

