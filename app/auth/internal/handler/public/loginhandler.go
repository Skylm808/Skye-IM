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

// 用户登录
func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err)
			return
		}

		l := public.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		if err != nil {
			response.Error(w, err)
			return
		}

		response.Success(w, resp)
	}
}
