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

// 更新群组已读进度
func UpdateGroupReadSeqHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateGroupReadSeqReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := membermgmt.NewUpdateGroupReadSeqLogic(r.Context(), svcCtx)
		resp, err := l.UpdateGroupReadSeq(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
