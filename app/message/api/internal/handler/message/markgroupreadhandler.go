// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package message

import (
	"net/http"

	"SkyeIM/app/message/api/internal/logic/message"
	"SkyeIM/app/message/api/internal/svc"
	"SkyeIM/app/message/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 群聊已读上报（更新readSeq）
func MarkGroupReadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MarkGroupReadReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := message.NewMarkGroupReadLogic(r.Context(), svcCtx)
		resp, err := l.MarkGroupRead(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
