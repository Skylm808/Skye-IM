// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package search

import (
	"net/http"

	"SkyeIM/app/user/api/internal/logic/search"
	"SkyeIM/app/user/api/internal/svc"
	"SkyeIM/app/user/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 全局模糊搜索（用户/群组）
func GlobalSearchHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GlobalSearchRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := search.NewGlobalSearchLogic(r.Context(), svcCtx)
		resp, err := l.GlobalSearch(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
