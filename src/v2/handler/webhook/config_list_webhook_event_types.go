package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/bb-consent/api/src/config"
	wh "github.com/bb-consent/api/src/v2/webhook"
)

// WebhookEventTypesResp Define response structure for webhook event types
type WebhookEventTypesResp struct {
	EventTypes []string `json:"eventTypes"`
}

// ConfigListWebhookEventTypes List available webhook event types
func ConfigListWebhookEventTypes(w http.ResponseWriter, r *http.Request) {

	resp := WebhookEventTypesResp{
		wh.WebhooksConfiguration.Events,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
