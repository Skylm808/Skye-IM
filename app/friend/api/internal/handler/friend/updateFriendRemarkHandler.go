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

// 更新好友备注
func UpdateFriendRemarkHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateRemarkReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := friend.NewUpdateFriendRemarkLogic(r.Context(), svcCtx)
		resp, err := l.UpdateFriendRemark(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
