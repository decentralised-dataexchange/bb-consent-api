package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/actionlog"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/individual"
	wh "github.com/bb-consent/api/internal/webhook"
	"github.com/bb-consent/api/internal/webhook_dispatcher"
	"github.com/gorilla/mux"
)

// ConfigRedeliverWebhookPayloadByDeliveryID Redo payload delivery to the webhook
func ConfigRedeliverWebhookPayloadByDeliveryID(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	webhookId := mux.Vars(r)[config.WebhookId]
	webhookId = common.Sanitize(webhookId)

	deliveryId := mux.Vars(r)[config.DeliveryId]
	deliveryId = common.Sanitize(deliveryId)

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

	// Validating the given delivery ID for a webhook
	webhookDelivery, err := wh.GetWebhookDeliveryByID(webhook.ID, deliveryId)
	if err != nil {
		m := fmt.Sprintf("Failed to get delivery details by ID:%v for webhook:%v for organisation: %v", deliveryId, webhookId, organisationId)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Converting the webhook payload to bytes
	webhookPayloadBytes, err := json.Marshal(webhookDelivery.RequestPayload)
	if err != nil {
		m := fmt.Sprintf("Failed to convert webhook event data to bytes, error:%v; Failed to redeliver payload for webhook for event:<%s>, user:<%s>, org:<%s>", err.Error(), webhookDelivery.WebhookEventType, webhookDelivery.UserID, organisationId)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	// Get the details of who triggered webhook
	u, err := individualRepo.Get(webhookDelivery.UserID)
	if err != nil {
		m := fmt.Sprintf("Failed to get user, error:%v; Failed to redeliver payload for webhook for event:<%s>, user:<%s>, org:<%s>", err.Error(), webhookDelivery.WebhookEventType, webhookDelivery.UserID, organisationId)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	go webhook_dispatcher.ProcessWebhooks(webhookDelivery.WebhookEventType, webhookPayloadBytes)

	// Log webhook calls in webhooks category
	aLog := fmt.Sprintf("Organization webhook: %v triggered by user: %v by event: %v", webhook.PayloadURL, u.Email, webhookDelivery.WebhookEventType)
	actionlog.LogOrgWebhookCalls(u.Id, u.Email, organisationId, aLog)

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

}
