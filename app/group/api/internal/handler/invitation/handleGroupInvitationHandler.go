// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package invitation

import (
	"net/http"

	"SkyeIM/app/group/api/internal/logic/invitation"
	"SkyeIM/app/group/api/internal/svc"
	"SkyeIM/app/group/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 处理群聊邀请
func HandleGroupInvitationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HandleGroupInvitationReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := invitation.NewHandleGroupInvitationLogic(r.Context(), svcCtx)
		resp, err := l.HandleGroupInvitation(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
