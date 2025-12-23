// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"net/http"

	"SkyeIM/common/response"
	"auth/internal/logic/user"
	"auth/internal/svc"
	"auth/internal/types"
)

// 退出登录
func LogoutHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewLogoutLogic(r.Context(), svcCtx)
		resp, err := l.Logout(&types.Empty{})
		if err != nil {
			response.Error(w, err)
			return
		}

		response.Success(w, resp)
	}
}
