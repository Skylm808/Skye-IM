// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package membermgmt

import (
	"net/http"

	"SkyeIM/app/group/api/internal/logic/membermgmt"
	"SkyeIM/app/group/api/internal/svc"
	"SkyeIM/app/group/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 设置成员禁言
func SetMemberMuteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SetMemberMuteReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := membermgmt.NewSetMemberMuteLogic(r.Context(), svcCtx)
		resp, err := l.SetMemberMute(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
