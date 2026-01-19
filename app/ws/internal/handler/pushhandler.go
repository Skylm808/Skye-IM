package handler

// pushhandler.go - 统一的推送接口
//
// 角色：内部推送接口
// 职责：
// 1. 接收来自各个 RPC 服务的推送请求（用户通知、群组事件）
// 2. 内部鉴权：校验 X-Skyeim-Push-Secret 头
// 3. 根据推送类型调用 Hub 的相应方法

import (
	"encoding/json"
	"net/http"

	"SkyeIM/app/ws/internal/conn"
	"SkyeIM/app/ws/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PushHandler struct {
	svcCtx *svc.ServiceContext
	hub    *conn.Hub
}

func NewPushHandler(svcCtx *svc.ServiceContext, hub *conn.Hub) *PushHandler {
	return &PushHandler{
		svcCtx: svcCtx,
		hub:    hub,
	}
}

type PushRequest struct {
	Type             string                 `json:"type"`             // "user" 或 "group"
	UserId           int64                  `json:"userId"`           // type=user 时使用
	NotificationType string                 `json:"notificationType"` // type=user 时使用
	GroupId          string                 `json:"groupId"`          // type=group 时使用
	EventType        string                 `json:"eventType"`        // type=group 时使用
	EventData        map[string]interface{} `json:"eventData"`        // type=group 时使用
	Data             map[string]interface{} `json:"data"`             // type=user 时使用
}

type PushResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (h *PushHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 内部接口鉴权
	if secret := h.svcCtx.Config.PushEvent.Secret; secret != "" {
		if r.Header.Get("X-Skyeim-Push-Secret") != secret {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(PushResponse{
				Code:    401,
				Message: "Unauthorized",
			})
			return
		}
	}

	var req PushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logx.Errorf("[Push] Decode error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PushResponse{
			Code:    400,
			Message: "Invalid request",
		})
		return
	}

	// 根据类型分发推送
	switch req.Type {
	case "user":
		h.handleUserPush(req)
	case "group":
		h.handleGroupPush(req)
	default:
		logx.Errorf("[Push] Unknown type: %s", req.Type)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PushResponse{
			Code:    400,
			Message: "Unknown push type",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PushResponse{
		Code:    0,
		Message: "success",
	})
}

// handleUserPush 处理用户推送
func (h *PushHandler) handleUserPush(req PushRequest) {
	message := &conn.Message{
		Type: req.NotificationType,
		Data: mustMarshal(req.Data),
	}

	success := h.hub.SendToUser(req.UserId, message)

	if success {
		logx.Infof("[Push] Pushed %s notification to user %d", req.NotificationType, req.UserId)
	} else {
		logx.Infof("[Push] User %d offline, notification not delivered", req.UserId)
	}
}

// handleGroupPush 处理群组推送
func (h *PushHandler) handleGroupPush(req PushRequest) {
	h.hub.NotifyGroupEvent(req.GroupId, req.EventType, req.EventData)
	logx.Infof("[Push] Pushed %s event to group %s", req.EventType, req.GroupId)
}