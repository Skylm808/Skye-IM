// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package friend

import (
	"net/http"

	"SkyeIM/app/friend/api/internal/logic/friend"
	"SkyeIM/app/friend/api/internal/svc"
	"SkyeIM/app/friend/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 删除好友
func DeleteFriendHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DeleteFriendReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := friend.NewDeleteFriendLogic(r.Context(), svcCtx)
		resp, err := l.DeleteFriend(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
