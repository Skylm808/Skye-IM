package handler

// pusheventhandler.go - 内部事件推送接口 (后台广播站)
//
// 角色：内部广播员 / 系统通知接口
// 职责：
// 1. 接收 HTTP 请求：通常由其他微服务（如 Group RPC）或后台管理系统发起。
// 2. 内部鉴权：校验 X-Skyeim-Push-Secret 头，防止接口被恶意调用。
// 3. 事件广播：调用 Hub 的方法，将事件（如"群解散"、"被踢出群"）推送到 WebSocket 连接。
//
// 关系说明：
// - 它不处理终端用户的 WebSocket 连接。
// - 它是 HTTP 到 WebSocket 的桥梁，允许外部系统主动触发 WebSocket 推送。
// - 它是 `wshandler` 的补充，前者处理用户行为，后者处理系统行为。

import (
	"encoding/json"
	"net/http"

	"SkyeIM/app/ws/internal/conn"
	"SkyeIM/app/ws/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PushEventHandler struct {
	svcCtx *svc.ServiceContext
	hub    *conn.Hub
}

func NewPushEventHandler(svcCtx *svc.ServiceContext, hub *conn.Hub) *PushEventHandler {
	return &PushEventHandler{
		svcCtx: svcCtx,
		hub:    hub,
	}
}

type PushEventRequest struct {
	GroupId   string                 `json:"groupId"`
	EventType string                 `json:"eventType"`
	EventData map[string]interface{} `json:"eventData"`
}

type PushEventResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (h *PushEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 内部接口鉴权（可选）
	if secret := h.svcCtx.Config.PushEvent.Secret; secret != "" {
		if r.Header.Get("X-Skyeim-Push-Secret") != secret {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(PushEventResponse{
				Code:    401,
				Message: "Unauthorized",
			})
			return
		}
	}

	var req PushEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logx.Errorf("[PushEvent] Decode error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PushEventResponse{
			Code:    400,
			Message: "Invalid request",
		})
		return
	}

	// 推送事件到群组
	h.hub.NotifyGroupEvent(req.GroupId, req.EventType, req.EventData)

	logx.Infof("[PushEvent] Pushed %s event to group %s", req.EventType, req.GroupId)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PushEventResponse{
		Code:    0,
		Message: "success",
	})
}
