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

// 获取用户信息
func GetUserInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewGetUserInfoLogic(r.Context(), svcCtx)
		resp, err := l.GetUserInfo(&types.Empty{})
		if err != nil {
			response.Error(w, err)
			return
		}

		response.Success(w, resp)
	}
}
