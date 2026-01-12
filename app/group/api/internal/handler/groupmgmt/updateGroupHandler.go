// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package groupmgmt

import (
	"net/http"

	"SkyeIM/app/group/api/internal/logic/groupmgmt"
	"SkyeIM/app/group/api/internal/svc"
	"SkyeIM/app/group/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 更新群组信息
func UpdateGroupHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateGroupReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := groupmgmt.NewUpdateGroupLogic(r.Context(), svcCtx)
		resp, err := l.UpdateGroup(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
