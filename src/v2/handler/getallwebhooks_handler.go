package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	wh "github.com/bb-consent/api/src/webhooks"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebhookWithLastDeliveryStatus Defines webhook structure along with last delivery status
type WebhookWithLastDeliveryStatus struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty"` // Webhook ID
	PayloadURL            string             // Webhook payload URL
	Disabled              bool               // Disabled or not
	TimeStamp             string             // UTC timestamp
	IsLastDeliverySuccess bool               // Indicates whether last payload delivery to webhook was success or not
}

// GetAllWebhooks Gets all webhooks for an organisation
func GetAllWebhooks(w http.ResponseWriter, r *http.Request) {
	// Reading URL parameters
	organizationID := r.Header.Get(config.OrganizationId)

	// Validating the given organisation ID
	_, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	sanitizedOrgId := common.Sanitize(organizationID)

	// Fetching all the webhooks for an organisation
	webhooks, err := wh.GetAllWebhooksByOrgID(sanitizedOrgId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch webhooks for organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var updatedWebhooks []WebhookWithLastDeliveryStatus

	updatedWebhooks = make([]WebhookWithLastDeliveryStatus, 0)

	for _, webhook := range webhooks {

		// Fetching the last delivery to the webhook and retrieving the delivery status
		isLastDeliverySuccess := false
		lastDelivery, err := wh.GetLastWebhookDelivery(webhook.ID.Hex())
		if err != nil {
			// There is no payload delivery yet !
			isLastDeliverySuccess = true
		} else {
			// if the last payload delivery is completed and response status code is within 2XX range
			if lastDelivery.Status == wh.DeliveryStatus[wh.DeliveryStatusCompleted] {
				if (lastDelivery.ResponseStatusCode >= 200 && lastDelivery.ResponseStatusCode <= 208) || lastDelivery.ResponseStatusCode == 226 {
					isLastDeliverySuccess = true
				}
			}

		}

		updatedWebhook := WebhookWithLastDeliveryStatus{
			ID:                    webhook.ID,
			PayloadURL:            webhook.PayloadURL,
			Disabled:              webhook.Disabled,
			TimeStamp:             webhook.TimeStamp,
			IsLastDeliverySuccess: isLastDeliverySuccess,
		}

		updatedWebhooks = append(updatedWebhooks, updatedWebhook)
	}

	response, _ := json.Marshal(updatedWebhooks)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
