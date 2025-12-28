// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package blacklist

import (
	"net/http"

	"SkyeIM/app/friend/api/internal/logic/blacklist"
	"SkyeIM/app/friend/api/internal/svc"
	"SkyeIM/app/friend/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取黑名单列表
func GetBlacklistHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PageReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := blacklist.NewGetBlacklistLogic(r.Context(), svcCtx)
		resp, err := l.GetBlacklist(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
