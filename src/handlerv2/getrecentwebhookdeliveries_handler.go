package handlerv2

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

// recentWebhookDelivery Defines the structure for recent webhook delivery
type recentWebhookDelivery struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"` // Webhook delivery ID
	WebhookID          string             // Webhook ID
	ResponseStatusCode int                // HTTP response status code
	ResponseStatusStr  string             // HTTP response status string
	TimeStamp          string             // UTC timestamp when webhook execution started
	Status             string             // Status of webhook delivery for e.g. failed or completed
	StatusDescription  string             // Describe the status for e.g. Reason for failure
}

type recentWebhookDeliveryResp struct {
	WebhookDeliveries []recentWebhookDelivery
	Links             common.PaginationLinks
}

// GetRecentWebhookDeliveries Gets the recent webhook deliveries limited by `x` records
func GetRecentWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	// Reading the URL parameters
	organizationID := r.Header.Get(config.OrganizationId)
	webhookID := mux.Vars(r)[config.WebhookId]

	startID, limit := common.ParsePaginationQueryParameters(r)
	if limit == 0 {
		limit = 50
	}

	// Validating the given organisation ID
	_, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	sanitizedOrgId := common.Sanitize(organizationID)
	sanitizedWebhookId := common.Sanitize(webhookID)

	// Validating the given webhook ID for an organisation
	webhook, err := wh.GetByOrgID(sanitizedWebhookId, sanitizedOrgId)
	if err != nil {
		m := fmt.Sprintf("Failed to get webhook:%v for organisation: %v", webhookID, organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Get all the recent webhook deliveries
	recentWebhookDeliveries, lastID, err := wh.GetAllDeliveryByWebhookID(webhook.ID.Hex(), startID, limit)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch recent payload deliveries for webhook:%v for organisation: %v", webhookID, organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// Constructing the response
	var resp recentWebhookDeliveryResp

	resp.WebhookDeliveries = make([]recentWebhookDelivery, 0)

	for _, wd := range recentWebhookDeliveries {

		tempRecentWebhookDelivery := recentWebhookDelivery{
			ID:                 wd.ID,
			WebhookID:          wd.WebhookID,
			ResponseStatusCode: wd.ResponseStatusCode,
			ResponseStatusStr:  wd.ResponseStatusStr,
			TimeStamp:          wd.ExecutionStartTimeStamp,
			Status:             wd.Status,
			StatusDescription:  wd.StatusDescription,
		}

		resp.WebhookDeliveries = append(resp.WebhookDeliveries, tempRecentWebhookDelivery)
	}

	resp.Links = common.CreatePaginationLinks(r, startID, lastID, limit)

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
