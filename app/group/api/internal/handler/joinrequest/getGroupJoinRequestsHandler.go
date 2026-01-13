// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package joinrequest

import (
	"net/http"

	"SkyeIM/app/group/api/internal/logic/joinrequest"
	"SkyeIM/app/group/api/internal/svc"
	"SkyeIM/app/group/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取群组的入群申请列表
func GetGroupJoinRequestsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetJoinRequestsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := joinrequest.NewGetGroupJoinRequestsLogic(r.Context(), svcCtx)
		resp, err := l.GetGroupJoinRequests(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
