package handler

import (
	"encoding/json"
	"net/http"

	"github.com/bb-consent/api/src/config"
	wh "github.com/bb-consent/api/src/webhooks"
)

// WebhookPayloadContentTypesResp Defines response structure for webhook payload content types
type WebhookPayloadContentTypesResp struct {
	ContentTypes []string
}

// GetWebhookPayloadContentTypes List available webhook payload content types
func GetWebhookPayloadContentTypes(w http.ResponseWriter, r *http.Request) {
	var webhookPayloadContentTypesResp WebhookPayloadContentTypesResp

	for _, payloadContentTypes := range wh.PayloadContentTypes {
		webhookPayloadContentTypesResp.ContentTypes = append(webhookPayloadContentTypesResp.ContentTypes, payloadContentTypes)
	}

	response, _ := json.Marshal(webhookPayloadContentTypesResp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
