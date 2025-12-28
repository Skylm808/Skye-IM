// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package request

import (
	"net/http"

	"SkyeIM/app/friend/api/internal/logic/request"
	"SkyeIM/app/friend/api/internal/svc"
	"SkyeIM/app/friend/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 处理好友申请
func HandleFriendRequestHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HandleFriendRequestReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := request.NewHandleFriendRequestLogic(r.Context(), svcCtx)
		resp, err := l.HandleFriendRequest(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
