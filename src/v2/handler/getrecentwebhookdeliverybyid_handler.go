package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	wh "github.com/bb-consent/api/src/webhooks"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type webhookDeliveryResp struct {
	ID                      primitive.ObjectID  `bson:"_id,omitempty"`
	RequestHeaders          map[string][]string // HTTP headers posted to webhook endpoint
	RequestPayload          interface{}         // JSON payload posted to webhook endpoint
	ResponseHeaders         map[string][]string // HTTP response headers received from webhook endpoint
	ResponseBody            string              // HTTP response body received from webhook endpoint in bytes
	ResponseStatusCode      int                 // HTTP response status code
	ResponseStatusStr       string              // HTTP response status string
	ExecutionStartTimeStamp string              // UTC timestamp when webhook execution started
	ExecutionEndTimeStamp   string              // UTC timestamp when webhook execution ended
	Status                  string              // Status of webhook delivery for e.g. failed or completed
	StatusDescription       string              // Describe the status for e.g. Reason for failure
}

// GetRecentWebhookDeliveryById Gets the payload delivery details for a webhook by ID
func GetRecentWebhookDeliveryById(w http.ResponseWriter, r *http.Request) {
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

	// Get the webhook delivery by ID
	webhookDelivery, err := wh.GetWebhookDeliveryByID(webhook.ID.Hex(), sanitizedDeliveryId)
	if err != nil {
		m := fmt.Sprintf("Failed to get delivery details by ID:%v for webhook:%v for organisation: %v", deliveryID, webhookID, organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	resp := webhookDeliveryResp{
		ID:                      webhookDelivery.ID,
		RequestHeaders:          webhookDelivery.RequestHeaders,
		RequestPayload:          webhookDelivery.RequestPayload,
		ResponseHeaders:         webhookDelivery.ResponseHeaders,
		ResponseBody:            webhookDelivery.ResponseBody,
		ResponseStatusCode:      webhookDelivery.ResponseStatusCode,
		ResponseStatusStr:       webhookDelivery.ResponseStatusStr,
		ExecutionStartTimeStamp: webhookDelivery.ExecutionStartTimeStamp,
		ExecutionEndTimeStamp:   webhookDelivery.ExecutionEndTimeStamp,
		Status:                  webhookDelivery.Status,
		StatusDescription:       webhookDelivery.StatusDescription,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
