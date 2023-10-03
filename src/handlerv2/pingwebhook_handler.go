package handlerv2

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	wh "github.com/bb-consent/api/src/webhooks"
	"github.com/gorilla/mux"
)

// PingWebhookResp Defines the response structure for webhook status check using ping
type PingWebhookResp struct {
	ResponseStatusCode      int    // HTTP response status code
	ResponseStatusStr       string // HTTP response status string
	ExecutionStartTimeStamp string // UTC timestamp when webhook execution started
	ExecutionEndTimeStamp   string // UTC timestamp when webhook execution ended
	Status                  string // Status of webhook delivery for e.g. failed or completed
	StatusDescription       string // Describe the status for e.g. Reason for failure
}

// PingWebhook Pings webhook payload URL to check the response status code is 200 OK or not
func PingWebhook(w http.ResponseWriter, r *http.Request) {

	// Reading the URL parameters
	organizationID := r.Header.Get(config.OrganizationId)
	webhookID := mux.Vars(r)[config.WebhookId]

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

	// Pinging webhook payload URL
	_, resp, executionStartTimeStamp, executionEndTimeStamp, err := wh.PingWebhook(webhook)

	if err != nil {

		log.Printf("Error: %v; Failed to ping webhook:%v for organisation: %v", err, webhookID, organizationID)

		// Constructing webhook ping response
		pingWebhookResp := PingWebhookResp{
			ExecutionStartTimeStamp: executionStartTimeStamp,
			ExecutionEndTimeStamp:   executionEndTimeStamp,
			Status:                  wh.DeliveryStatus[wh.DeliveryStatusFailed],
			StatusDescription:       err.Error(),
		}

		response, _ := json.Marshal(pingWebhookResp)
		w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(response)
		return
	}

	defer resp.Body.Close()

	// Constructing webhook ping response
	pingWebhookResp := PingWebhookResp{
		ResponseStatusCode:      resp.StatusCode,
		ResponseStatusStr:       resp.Status,
		ExecutionStartTimeStamp: executionStartTimeStamp,
		ExecutionEndTimeStamp:   executionEndTimeStamp,
		Status:                  wh.DeliveryStatus[wh.DeliveryStatusCompleted],
	}

	response, _ := json.Marshal(pingWebhookResp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
