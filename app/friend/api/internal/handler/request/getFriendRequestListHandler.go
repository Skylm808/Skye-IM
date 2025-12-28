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

// 获取收到的好友申请列表
func GetFriendRequestListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PageReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := request.NewGetFriendRequestListLogic(r.Context(), svcCtx)
		resp, err := l.GetFriendRequestList(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
