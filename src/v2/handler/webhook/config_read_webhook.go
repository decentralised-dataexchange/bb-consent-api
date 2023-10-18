package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	wh "github.com/bb-consent/api/src/v2/webhook"
	"github.com/gorilla/mux"
)

type readWebhookResp struct {
	Webhook wh.Webhook `json:"webhook"`
}

// ConfigReadWebhook
func ConfigReadWebhook(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	webhookId := mux.Vars(r)[config.WebhookId]
	webhookId = common.Sanitize(webhookId)

	// Repository
	webhookRepo := wh.WebhookRepository{}
	webhookRepo.Init(organisationId)

	// Fetching webhook by ID
	webhook, err := webhookRepo.GetByOrgID(webhookId)
	if err != nil {
		m := fmt.Sprintf("Failed to get webhook:%v for organisation: %v", webhookId, organisationId)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}
	resp := readWebhookResp{
		Webhook: webhook,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
