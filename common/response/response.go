package response

import (
	"net/http"

	"SkyeIM/common/errorx"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success 成功响应
func Success(w http.ResponseWriter, data interface{}) {
	resp := &Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
	httpx.OkJson(w, resp)
}

// Error 错误响应
func Error(w http.ResponseWriter, err error) {
	code := errorx.CodeUnknown
	msg := err.Error()

	// 判断是否为业务错误
	if e, ok := err.(*errorx.CodeError); ok {
		code = e.GetCode()
		msg = e.GetMessage()
	}

	resp := &Response{
		Code:    code,
		Message: msg,
		Data:    nil,
	}
	httpx.OkJson(w, resp)
}

// ParamError 参数错误响应
func ParamError(w http.ResponseWriter, err error) {
	resp := &Response{
		Code:    errorx.CodeParam,
		Message: err.Error(),
		Data:    nil,
	}
	httpx.OkJson(w, resp)
}
