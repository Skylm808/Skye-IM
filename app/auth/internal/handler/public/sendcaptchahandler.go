// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package public

import (
	"net/http"

	"SkyeIM/common/response"
	"auth/internal/logic/public"
	"auth/internal/svc"
	"auth/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 发送邮箱验证码
func SendCaptchaHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendCaptchaRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err)
			return
		}

		l := public.NewSendCaptchaLogic(r.Context(), svcCtx)
		resp, err := l.SendCaptcha(&req)
		if err != nil {
			response.Error(w, err)
			return
		}

		response.Success(w, resp)
	}
}
