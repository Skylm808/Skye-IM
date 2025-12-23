// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"net/http"

	"SkyeIM/common/response"
	"auth/internal/logic/user"
	"auth/internal/svc"
	"auth/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 修改密码
func ChangePasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ChangePasswordRequest
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err)
			return
		}

		l := user.NewChangePasswordLogic(r.Context(), svcCtx)
		resp, err := l.ChangePassword(&req)
		if err != nil {
			response.Error(w, err)
			return
		}

		response.Success(w, resp)
	}
}
