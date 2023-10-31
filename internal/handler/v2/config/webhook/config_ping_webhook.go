package webhook

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	wh "github.com/bb-consent/api/internal/webhook"
	"github.com/gorilla/mux"
)

// PingWebhookResp Defines the response structure for webhook status check using ping
type PingWebhookResp struct {
	ResponseStatusCode      int    `json:"responseStatusCode"`      // HTTP response status code
	ResponseStatusStr       string `json:"responseStatusStr"`       // HTTP response status string
	ExecutionStartTimeStamp string `json:"executionStartTimestamp"` // UTC timestamp when webhook execution started
	ExecutionEndTimeStamp   string `json:"executionEndTimestamp"`   // UTC timestamp when webhook execution ended
	Status                  string `json:"status"`                  // Status of webhook delivery for e.g. failed or completed
	StatusDescription       string `json:"statusDescription"`       // Describe the status for e.g. Reason for failure
}

func ConfigPingWebhook(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	webhookId := mux.Vars(r)[config.WebhookId]
	webhookId = common.Sanitize(webhookId)

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

	// Pinging webhook payload URL
	_, resp, executionStartTimeStamp, executionEndTimeStamp, err := wh.PingWebhook(webhook)

	if err != nil {

		log.Printf("Error: %v; Failed to ping webhook:%v for organisation: %v", err, webhookId, organisationId)

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
