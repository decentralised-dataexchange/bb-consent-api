package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/actionlog"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/user"
	wh "github.com/bb-consent/api/src/webhooks"
	"github.com/gorilla/mux"
)

// RedeliverWebhookPayloadByDeliveryID Redo payload delivery to the webhook
func RedeliverWebhookPayloadByDeliveryID(w http.ResponseWriter, r *http.Request) {
	// Reading the URL parameters
	organizationID := r.Header.Get(config.OrganizationId)
	webhookID := mux.Vars(r)[config.WebhookId]
	deliveryID := mux.Vars(r)[config.WebhookDeliveryId]

	// Validating the given organisation ID
	_, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	sanitizedOrgId := common.Sanitize(organizationID)
	sanitizedWebhookId := common.Sanitize(webhookID)
	sanitizedDeliveryId := common.Sanitize(deliveryID)

	// Validating the given webhook ID for an organisation
	webhook, err := wh.GetByOrgID(sanitizedWebhookId, sanitizedOrgId)
	if err != nil {
		m := fmt.Sprintf("Failed to get webhook:%v for organisation: %v", webhookID, organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Validating the given delivery ID for a webhook
	webhookDelivery, err := wh.GetWebhookDeliveryByID(webhook.ID.Hex(), sanitizedDeliveryId)
	if err != nil {
		m := fmt.Sprintf("Failed to get delivery details by ID:%v for webhook:%v for organisation: %v", deliveryID, webhookID, organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Converting the webhook payload to bytes
	webhookPayloadBytes, err := json.Marshal(webhookDelivery.RequestPayload)
	if err != nil {
		m := fmt.Sprintf("Failed to convert webhook event data to bytes, error:%v; Failed to redeliver payload for webhook for event:<%s>, user:<%s>, org:<%s>", err.Error(), webhookDelivery.WebhookEventType, webhookDelivery.UserID, organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// Get the details of who triggered webhook
	u, err := user.Get(webhookDelivery.UserID)
	if err != nil {
		m := fmt.Sprintf("Failed to get user, error:%v; Failed to redeliver payload for webhook for event:<%s>, user:<%s>, org:<%s>", err.Error(), webhookDelivery.WebhookEventType, webhookDelivery.UserID, organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	go wh.PushWebhookEventToKafkaTopic(webhookDelivery.WebhookEventType, webhookPayloadBytes, wh.WebhooksConfiguration.KafkaConfig.Topic)

	// Log webhook calls in webhooks category
	aLog := fmt.Sprintf("Organization webhook: %v triggered by user: %v by event: %v", webhook.PayloadURL, u.Email, webhookDelivery.WebhookEventType)
	actionlog.LogOrgWebhookCalls(u.ID.Hex(), u.Email, organizationID, aLog)

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

}
