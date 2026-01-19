package wspush

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// WsPushClient WebSocket 推送客户端（公共）
// 用于从各个 RPC 服务向 WebSocket 服务推送通知
type WsPushClient struct {
	wsServiceUrl string
	secret       string
	httpClient   *http.Client
}

// NewWsPushClient 创建推送客户端
func NewWsPushClient(wsServiceUrl string, secret string) *WsPushClient {
	return &WsPushClient{
		wsServiceUrl: wsServiceUrl,
		secret:       secret,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

// PushToUser 推送通知给指定用户
// 用于好友请求、群组邀请等单用户通知
func (c *WsPushClient) PushToUser(userId int64, notificationType string, data map[string]interface{}) error {
	reqData := map[string]interface{}{
		"type":             "user",
		"userId":           userId,
		"notificationType": notificationType,
		"data":             data,
	}

	return c.sendPushRequest(reqData, fmt.Sprintf("user notification %s to user %d", notificationType, userId))
}

// PushGroupEvent 推送群组事件
// 用于群组解散、成员踢出、成员加入等群组事件
func (c *WsPushClient) PushGroupEvent(groupId, eventType string, eventData map[string]interface{}) error {
	reqData := map[string]interface{}{
		"type":      "group",
		"groupId":   groupId,
		"eventType": eventType,
		"eventData": eventData,
	}

	return c.sendPushRequest(reqData, fmt.Sprintf("group event %s to group %s", eventType, groupId))
}

// sendPushRequest 发送推送请求的内部实现
func (c *WsPushClient) sendPushRequest(reqData map[string]interface{}, logDesc string) error {
	jsonData, _ := json.Marshal(reqData)

	url := fmt.Sprintf("%s/api/push", c.wsServiceUrl)
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.secret != "" {
		httpReq.Header.Set("X-Skyeim-Push-Secret", c.secret)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logx.Errorf("[WsPush] HTTP request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logx.Errorf("[WsPush] HTTP status: %d", resp.StatusCode)
		return fmt.Errorf("push failed with status %d", resp.StatusCode)
	}

	logx.Infof("[WsPush] Pushed %s", logDesc)
	return nil
}
