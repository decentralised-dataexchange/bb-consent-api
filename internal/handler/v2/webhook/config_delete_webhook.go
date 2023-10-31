package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	wh "github.com/bb-consent/api/internal/webhook"
	"github.com/gorilla/mux"
)

type deleteWebhookResp struct {
	Webhook wh.Webhook `json:"webhook"`
}

// ConfigDeleteWebhook
func ConfigDeleteWebhook(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	webhookId := mux.Vars(r)[config.WebhookId]
	webhookId = common.Sanitize(webhookId)

	// Repository
	webhookRepo := wh.WebhookRepository{}
	webhookRepo.Init(organisationId)

	// Fetching webhook by ID
	toBeDeletedWebhook, err := webhookRepo.GetByOrgID(webhookId)
	if err != nil {
		m := fmt.Sprintf("Failed to get webhook:%v for organisation: %v", webhookId, organisationId)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	toBeDeletedWebhook.IsDeleted = true

	// Save to db
	savedWebhook, err := webhookRepo.UpdateWebhook(toBeDeletedWebhook)
	if err != nil {
		m := fmt.Sprintf("Failed to update webhook:%v for organisation: %v", webhookId, organisationId)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := deleteWebhookResp{
		Webhook: savedWebhook,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
