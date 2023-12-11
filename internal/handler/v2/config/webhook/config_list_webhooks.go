package webhook

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/paginate"
	wh "github.com/bb-consent/api/internal/webhook"
)

// WebhookWithLastDeliveryStatus Defines webhook structure along with last delivery status
type WebhookWithLastDeliveryStatus struct {
	ID                    string   `json:"id" bson:"_id,omitempty"` // Webhook ID
	OrganisationId        string   `json:"orgId" bson:"orgid"`
	PayloadURL            string   `json:"payloadUrl"`            // Webhook payload URL
	ContentType           string   `json:"contentType"`           // Webhook payload content type for e.g application/json
	SubscribedEvents      []string `json:"subscribedEvents"`      // Events subscribed for e.g. user.data.delete
	Disabled              bool     `json:"disabled"`              // Disabled or not
	SecretKey             string   `json:"secretKey"`             // For calculating SHA256 HMAC to verify data integrity and authenticity
	SkipSSLVerification   bool     `json:"skipSslVerification"`   // Skip SSL certificate verification or not (expiry is checked)
	TimeStamp             string   `json:"timestamp"`             // UTC timestamp
	IsLastDeliverySuccess bool     `json:"isLastDeliverySuccess"` // Indicates whether last payload delivery to webhook was success or not
}

func webhooksToInterfaceSlice(webhooks []WebhookWithLastDeliveryStatus) []interface{} {
	interfaceSlice := make([]interface{}, len(webhooks))
	for i, r := range webhooks {
		interfaceSlice[i] = r
	}
	return interfaceSlice
}

type listWebhooksResp struct {
	Webhooks   interface{}         `json:"webhooks"`
	Pagination paginate.Pagination `json:"pagination"`
}

// ConfigListWebhooks
func ConfigListWebhooks(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Repository
	webhookRepo := wh.WebhookRepository{}
	webhookRepo.Init(organisationId)

	// Fetching all the webhooks for an organisation
	webhooks, err := webhookRepo.GetAllWebhooksByOrgID()
	if err != nil {
		m := fmt.Sprintf("Failed to fetch webhooks for organization: %v", organisationId)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var updatedWebhooks []WebhookWithLastDeliveryStatus

	updatedWebhooks = make([]WebhookWithLastDeliveryStatus, 0)

	for _, webhook := range webhooks {

		// Fetching the last delivery to the webhook and retrieving the delivery status
		isLastDeliverySuccess := false
		lastDelivery, err := wh.GetLastWebhookDelivery(webhook.ID)
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
			OrganisationId:        webhook.OrganisationId,
			PayloadURL:            webhook.PayloadURL,
			ContentType:           webhook.ContentType,
			SubscribedEvents:      webhook.SubscribedEvents,
			Disabled:              webhook.Disabled,
			SecretKey:             webhook.SecretKey,
			SkipSSLVerification:   webhook.SkipSSLVerification,
			TimeStamp:             webhook.TimeStamp,
			IsLastDeliverySuccess: isLastDeliverySuccess,
		}

		updatedWebhooks = append(updatedWebhooks, updatedWebhook)
	}

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

	query := paginate.PaginateObjectsQuery{
		Limit:  limit,
		Offset: offset,
	}

	interfaceSlice := webhooksToInterfaceSlice(updatedWebhooks)
	result := paginate.PaginateObjects(query, interfaceSlice)
	var resp = listWebhooksResp{
		Webhooks:   result.Items,
		Pagination: result.Pagination,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
