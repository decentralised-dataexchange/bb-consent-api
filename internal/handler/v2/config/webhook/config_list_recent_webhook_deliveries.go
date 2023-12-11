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
	"github.com/gorilla/mux"
)

// recentWebhookDelivery Defines the structure for recent webhook delivery
type recentWebhookDelivery struct {
	Id                 string `json:"id" bson:"_id,omitempty"` // Webhook delivery ID
	WebhookId          string `json:"webhookId"`               // Webhook ID
	ResponseStatusCode int    `json:"responseStatusCode"`      // HTTP response status code
	ResponseStatusStr  string `json:"responseStatusStr"`       // HTTP response status string
	TimeStamp          string `json:"timestamp"`               // UTC timestamp when webhook execution started
	Status             string `json:"status"`                  // Status of webhook delivery for e.g. failed or completed
	StatusDescription  string `json:"statusDescription"`       // Describe the status for e.g. Reason for failure
}

func webhookDeliveriesToInterfaceSlice(webhookdeliveries []recentWebhookDelivery) []interface{} {
	interfaceSlice := make([]interface{}, len(webhookdeliveries))
	for i, r := range webhookdeliveries {
		interfaceSlice[i] = r
	}
	return interfaceSlice
}

type listWebhookDeliveriesResp struct {
	WebhookDeliveries interface{}         `json:"webhookDeliveries"`
	Pagination        paginate.Pagination `json:"pagination"`
}

// ConfigListRecentWebhookDeliveries
func ConfigListRecentWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	webhookId := mux.Vars(r)[config.WebhookId]
	webhookId = common.Sanitize(webhookId)

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

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

	// Get all recent webhook deliveries
	recentWebhookDeliveries, err := wh.GetAllDeliveryByWebhookId(webhook.ID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch recent payload deliveries for webhook:%v for organisation: %v", webhookId, organisationId)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// Constructing the response
	var webhookDeliveries []recentWebhookDelivery

	for _, wd := range recentWebhookDeliveries {

		tempRecentWebhookDelivery := recentWebhookDelivery{
			Id:                 wd.ID,
			WebhookId:          wd.WebhookID,
			ResponseStatusCode: wd.ResponseStatusCode,
			ResponseStatusStr:  wd.ResponseStatusStr,
			TimeStamp:          wd.ExecutionStartTimeStamp,
			Status:             wd.Status,
			StatusDescription:  wd.StatusDescription,
		}

		webhookDeliveries = append(webhookDeliveries, tempRecentWebhookDelivery)
	}

	query := paginate.PaginateObjectsQuery{
		Limit:  limit,
		Offset: offset,
	}

	interfaceSlice := webhookDeliveriesToInterfaceSlice(webhookDeliveries)
	result := paginate.PaginateObjects(query, interfaceSlice)
	var resp = listWebhookDeliveriesResp{
		WebhookDeliveries: result.Items,
		Pagination:        result.Pagination,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
