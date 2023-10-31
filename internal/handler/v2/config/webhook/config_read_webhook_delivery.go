package webhook

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	wh "github.com/bb-consent/api/internal/webhook"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type readWebhookDeliveryResp struct {
	Id                 primitive.ObjectID `json:"id" bson:"_id,omitempty"` // Webhook delivery ID
	WebhookId          string             `json:"webhookId"`               // Webhook ID
	ResponseStatusCode int                `json:"responseStatusCode"`      // HTTP response status code
	ResponseStatusStr  string             `json:"responseStatusStr"`       // HTTP response status string
	TimeStamp          string             `json:"timestamp"`               // UTC timestamp when webhook execution started
	Status             string             `json:"status"`                  // Status of webhook delivery for e.g. failed or completed
	StatusDescription  string             `json:"statusDescription"`       // Describe the status for e.g. Reason for failure
}

// GetRecentWebhookDeliveryById Gets the payload delivery details for a webhook by ID
func ConfigReadRecentWebhookDelivery(w http.ResponseWriter, r *http.Request) {
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

	// Get the webhook delivery by ID
	webhookDelivery, err := wh.GetWebhookDeliveryByID(webhook.ID.Hex(), deliveryId)
	if err != nil {
		m := fmt.Sprintf("Failed to get delivery details by ID:%v for webhook:%v for organisation: %v", deliveryId, webhookId, organisationId)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	resp := readWebhookDeliveryResp{
		Id:                 webhookDelivery.ID,
		WebhookId:          webhookDelivery.WebhookID,
		ResponseStatusCode: webhookDelivery.ResponseStatusCode,
		ResponseStatusStr:  webhookDelivery.ResponseStatusStr,
		TimeStamp:          webhookDelivery.ExecutionStartTimeStamp,
		Status:             webhookDelivery.Status,
		StatusDescription:  webhookDelivery.StatusDescription,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
