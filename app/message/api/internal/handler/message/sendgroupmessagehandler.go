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

// 发送群聊消息（可选：主要走WS）
func SendGroupMessageHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SendGroupMessageReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := message.NewSendGroupMessageLogic(r.Context(), svcCtx)
		resp, err := l.SendGroupMessage(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
