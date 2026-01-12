package handler

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
