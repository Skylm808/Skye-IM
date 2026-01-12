package wspush

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// WsPushClient WebSocket 推送客户端
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

// PushGroupEvent 推送群组事件
func (c *WsPushClient) PushGroupEvent(groupId, eventType string, eventData map[string]interface{}) error {
	reqData := map[string]interface{}{
		"groupId":   groupId,
		"eventType": eventType,
		"eventData": eventData,
	}

	jsonData, _ := json.Marshal(reqData)

	url := fmt.Sprintf("%s/api/push/event", c.wsServiceUrl)
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

	logx.Infof("[WsPush] Pushed %s event to group %s", eventType, groupId)
	return nil
}
