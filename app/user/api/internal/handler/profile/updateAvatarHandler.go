// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package profile

import (
	"net/http"

	"SkyeIM/app/user/api/internal/logic/profile"
	"SkyeIM/app/user/api/internal/svc"
	"SkyeIM/app/user/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 更新头像
func UpdateAvatarHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateAvatarRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := profile.NewUpdateAvatarLogic(r.Context(), svcCtx)
		resp, err := l.UpdateAvatar(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
