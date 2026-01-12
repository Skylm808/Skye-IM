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

// 获取收到的群聊邀请
func GetReceivedInvitationsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetInvitationsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := invitation.NewGetReceivedInvitationsLogic(r.Context(), svcCtx)
		resp, err := l.GetReceivedInvitations(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
