// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package profile

import (
	"net/http"

	"SkyeIM/app/user/api/internal/logic/profile"
	"SkyeIM/app/user/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取当前用户资料
func GetProfileHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := profile.NewGetProfileLogic(r.Context(), svcCtx)
		resp, err := l.GetProfile()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
