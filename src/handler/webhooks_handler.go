package handler

import (
	"encoding/json"
	"net/http"

	wh "github.com/bb-consent/api/src/webhooks"
)

// WebhookEventTypesResp Define response structure for webhook event types
type WebhookEventTypesResp struct {
	EventTypes []string
}

// GetWebhookEventTypes List available webhook event types
func GetWebhookEventTypes(w http.ResponseWriter, r *http.Request) {
	var webhookEventTypesResp WebhookEventTypesResp

	for _, eventType := range wh.EventTypes {
		webhookEventTypesResp.EventTypes = append(webhookEventTypesResp.EventTypes, eventType)
	}

	response, _ := json.Marshal(webhookEventTypesResp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
