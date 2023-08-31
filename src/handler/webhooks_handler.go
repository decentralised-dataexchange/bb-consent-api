package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/actionlog"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/user"
	wh "github.com/bb-consent/api/src/webhooks"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

// WebhookEventTypesResp Define response structure for webhook event types
type WebhookEventTypesResp struct {
	EventTypes []string
}

// GetWebhookEventTypes List available webhook event types
func GetWebhookEventTypes(w http.ResponseWriter, r *http.Request) {
	var webhookEventTypesResp WebhookEventTypesResp

	for _, eventType := range wh.EventTypes {
		webhookEventTypesResp.EventTypes = append(webhookEventTypesResp.EventTypes, eventType)
	}

	response, _ := json.Marshal(webhookEventTypesResp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}

// WebhookPayloadContentTypesResp Defines response structure for webhook payload content types
type WebhookPayloadContentTypesResp struct {
	ContentTypes []string
}

// GetWebhookPayloadContentTypes List available webhook payload content types
func GetWebhookPayloadContentTypes(w http.ResponseWriter, r *http.Request) {
	var webhookPayloadContentTypesResp WebhookPayloadContentTypesResp

	for _, payloadContentTypes := range wh.PayloadContentTypes {
		webhookPayloadContentTypesResp.ContentTypes = append(webhookPayloadContentTypesResp.ContentTypes, payloadContentTypes)
	}

	response, _ := json.Marshal(webhookPayloadContentTypesResp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// createWebhookReq Defines the request payload structure for creating a webhook for an organisation
type createWebhookReq struct {
	PayloadURL          string   `valid:"required,url"` // Webhook endpoint URL
	SubscribedEvents    []string `valid:"required"`     // Subscribed events
	ContentType         string   `valid:"required"`     // Data format for the webhook payload
	Disabled            bool     // Disabled or not
	SecretKey           string   // For calculating SHA256 HMAC to verify data integrity and authenticity
	SkipSSLVerification bool     // Skip SSL certificate verification or not (expiry is checked)
}

// uniqueSlice Filter out all the duplicate strings and returns the unique slice
func uniqueSlice(inputSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range inputSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// CreateWebhook Creates a webhook endpoint for an organisation
func CreateWebhook(w http.ResponseWriter, r *http.Request) {
	var requestPayload createWebhookReq

	// Reading request body as bytes and unmarshalling to a struct
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &requestPayload)

	organizationID := mux.Vars(r)["orgID"]

	// Validating request payload struct
	_, err := govalidator.ValidateStruct(requestPayload)
	if err != nil {
		log.Printf("Missing mandatory params; Failed to create webhook for organisation: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Validating the given organisation ID
	_, err = org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Check if the webhook endpoint contains http:// or https://
	if !(strings.HasPrefix(requestPayload.PayloadURL, "https://") || strings.HasPrefix(requestPayload.PayloadURL, "http://")) {
		m := fmt.Sprintf("Please prefix the endpoint URL with https:// or http://; Failed to create webhook for organisation:%v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Check if webhook with provided payload URL already exists
	count, err := wh.GetWebhookCountByPayloadURL(organizationID, requestPayload.PayloadURL)
	if err != nil {
		m := fmt.Sprintf("Failed to create webhook for organisation:%v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	if count > 0 {
		m := fmt.Sprintf("Webhook with provided payload URL already exists; Failed to create webhook for organisation:%v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Check if subscribed event type(s) array is empty
	if len(requestPayload.SubscribedEvents) == 0 {
		m := fmt.Sprintf("Provide atleast 1 event type; Failed to create webhook for organisation:%v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Check if subscribed event type(s) contains duplicates
	requestPayload.SubscribedEvents = uniqueSlice(requestPayload.SubscribedEvents)

	// Check the subscribed event type ID(s) provided is valid
	var isValidSubscribedEvents bool
	for _, subscribedEventType := range requestPayload.SubscribedEvents {
		isValidSubscribedEvents = false
		for _, eventType := range wh.EventTypes {
			if subscribedEventType == eventType {
				isValidSubscribedEvents = true
				break
			}
		}

		if !isValidSubscribedEvents {
			break
		}
	}

	if !isValidSubscribedEvents {
		m := fmt.Sprintf("Please provide a valid event type; Failed to create webhook for organisation:%v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Check if the content type ID provided is valid
	isValidContentType := false
	for _, payloadContentType := range wh.PayloadContentTypes {
		if requestPayload.ContentType == payloadContentType {
			isValidContentType = true
		}
	}

	if !isValidContentType {
		m := fmt.Sprintf("Please provide a valid content type; Failed to create webhook for organisation:%v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Constructing webhook structure required for adding it to database
	webhook := wh.Webhook{
		OrgID:               organizationID,
		PayloadURL:          strings.TrimSpace(requestPayload.PayloadURL),
		ContentType:         requestPayload.ContentType,
		SubscribedEvents:    requestPayload.SubscribedEvents,
		Disabled:            requestPayload.Disabled,
		SecretKey:           strings.TrimSpace(requestPayload.SecretKey),
		SkipSSLVerification: requestPayload.SkipSSLVerification,
		TimeStamp:           strconv.FormatInt(time.Now().UTC().Unix(), 10),
	}

	// Creating webhook
	webhook, err = wh.CreateWebhook(webhook)
	if err != nil {
		m := fmt.Sprintf("Failed to create webhook for organisation:%v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(webhook)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)

}

// WebhookWithLastDeliveryStatus Defines webhook structure along with last delivery status
type WebhookWithLastDeliveryStatus struct {
	ID                    bson.ObjectId `bson:"_id,omitempty"` // Webhook ID
	PayloadURL            string        // Webhook payload URL
	Disabled              bool          // Disabled or not
	TimeStamp             string        // UTC timestamp
	IsLastDeliverySuccess bool          // Indicates whether last payload delivery to webhook was success or not
}

// GetAllWebhooks Gets all webhooks for an organisation
func GetAllWebhooks(w http.ResponseWriter, r *http.Request) {
	// Reading URL parameters
	organizationID := mux.Vars(r)["orgID"]

	// Validating the given organisation ID
	_, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Fetching all the webhooks for an organisation
	webhooks, err := wh.GetAllWebhooksByOrgID(organizationID)
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// GetWebhook Gets a webhook for an organisation by ID
func GetWebhook(w http.ResponseWriter, r *http.Request) {
	// Reading URL parameters
	organizationID := mux.Vars(r)["orgID"]
	webhookID := mux.Vars(r)["webhookID"]

	// Validating the given organisation ID
	_, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Fetching webhook by ID for an organisation
	webhook, err := wh.GetByOrgID(webhookID, organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get webhook:%v for organisation: %v", webhookID, organizationID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	response, _ := json.Marshal(webhook)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// DeleteWebhook Deletes a webhook for an organisation
func DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	// Reading URL parameters
	organizationID := mux.Vars(r)["orgID"]
	webhookID := mux.Vars(r)["webhookID"]

	// Validating the given organisation ID
	_, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Validating the given webhook ID for an organisation
	_, err = wh.GetByOrgID(webhookID, organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get webhook:%v for organisation: %v", webhookID, organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Deleting webhook
	err = wh.DeleteWebhook(webhookID)
	if err != nil {
		m := fmt.Sprintf("Failed to delete webhook:%v for organisation: %v", webhookID, organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)

}

// updateWebhookReq Defines the request payload structure for updating a webhook for an organisation
type updateWebhookReq struct {
	PayloadURL          string   `valid:"required,url"` // Webhook endpoint URL
	SubscribedEvents    []string `valid:"required"`     // Events subscribed for e.g. user.data.delete
	ContentType         string   `valid:"required"`     // Data format for the webhook payload
	Disabled            bool     // Disabled or not
	SecretKey           string   // For calculating SHA256 HMAC to verify data integrity and authenticity
	SkipSSLVerification bool     // Skip SSL certificate verification or not (expiry is checked)
}

// UpdateWebhook Updates a webhook for an organisation by ID
func UpdateWebhook(w http.ResponseWriter, r *http.Request) {
	// Reading the URL parameters
	organizationID := mux.Vars(r)["orgID"]
	webhookID := mux.Vars(r)["webhookID"]

	fmt.Printf("Mux vars : %v\n", mux.Vars(r))

	// Validating the given organisation ID
	_, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Validating the given webhook ID for an organisation
	webhook, err := wh.GetByOrgID(webhookID, organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get webhook:%v for organisation: %v", webhookID, organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	var requestPayload updateWebhookReq

	// Reading request body as bytes and unmarshalling to a struct
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &requestPayload)

	// Validating request payload struct
	valid, err := govalidator.ValidateStruct(requestPayload)
	if valid != true {
		log.Printf("Missing mandatory params; Failed to update webhook for organisation: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Check if the webhook endpoint contains http:// or https://
	if !(strings.HasPrefix(requestPayload.PayloadURL, "https://") || strings.HasPrefix(requestPayload.PayloadURL, "http://")) {
		m := fmt.Sprintf("Please prefix the endpoint URL with https:// or http://; Failed to update webhook for organisation:%v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Check if webhook with provided payload URL already exists
	tempWebhook, err := wh.GetWebhookByPayloadURL(organizationID, requestPayload.PayloadURL)
	if err == nil {
		if tempWebhook.ID.Hex() != webhookID {
			m := fmt.Sprintf("Webhook with provided payload URL already exists; Failed to update webhook for organisation:%v", organizationID)
			common.HandleError(w, http.StatusBadRequest, m, err)
			return
		}
	}

	// Check if subscribed events array is empty
	if len(requestPayload.SubscribedEvents) == 0 {
		m := fmt.Sprintf("Provide atleast 1 event type in subscribed events; Failed to update webhook for organisation:%v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Check if subscribed event type(s) contains duplicates
	requestPayload.SubscribedEvents = uniqueSlice(requestPayload.SubscribedEvents)

	// Check the subscribed event type ID(s) provided is valid
	var isValidSubscribedEvents bool
	for _, subscribedEventType := range requestPayload.SubscribedEvents {
		isValidSubscribedEvents = false
		for _, eventType := range wh.EventTypes {
			if subscribedEventType == eventType {
				isValidSubscribedEvents = true
				break
			}
		}

		if !isValidSubscribedEvents {
			break
		}
	}

	if !isValidSubscribedEvents {
		m := fmt.Sprintf("Invalid event type provided in subscribed events; Failed to update webhook for organisation:%v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Check if the content type ID provided is valid
	isValidContentType := false
	for _, payloadContentType := range wh.PayloadContentTypes {
		if requestPayload.ContentType == payloadContentType {
			isValidContentType = true
		}
	}

	if !isValidContentType {
		m := fmt.Sprintf("Invalid content type provided; Failed to update webhook for organisation:%v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Updating webhook
	webhook.PayloadURL = strings.TrimSpace(requestPayload.PayloadURL)
	webhook.ContentType = requestPayload.ContentType
	webhook.SubscribedEvents = requestPayload.SubscribedEvents
	webhook.Disabled = requestPayload.Disabled
	webhook.SecretKey = requestPayload.SecretKey
	webhook.SkipSSLVerification = requestPayload.SkipSSLVerification

	webhook, err = wh.UpdateWebhook(webhook)
	if err != nil {
		m := fmt.Sprintf("Failed to update webhook:%v for organisation: %v", webhookID, organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(webhook)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}

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
	organizationID := mux.Vars(r)["orgID"]
	webhookID := mux.Vars(r)["webhookID"]

	// Validating the given organisation ID
	_, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Validating the given webhook ID for an organisation
	webhook, err := wh.GetByOrgID(webhookID, organizationID)
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
		w.Header().Set("Content-Type", "application/json")
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}

// recentWebhookDelivery Defines the structure for recent webhook delivery
type recentWebhookDelivery struct {
	ID                 bson.ObjectId `bson:"_id,omitempty"` // Webhook delivery ID
	WebhookID          string        // Webhook ID
	ResponseStatusCode int           // HTTP response status code
	ResponseStatusStr  string        // HTTP response status string
	TimeStamp          string        // UTC timestamp when webhook execution started
	Status             string        // Status of webhook delivery for e.g. failed or completed
	StatusDescription  string        // Describe the status for e.g. Reason for failure
}

type recentWebhookDeliveryResp struct {
	WebhookDeliveries []recentWebhookDelivery
	Links             common.PaginationLinks
}

// GetRecentWebhookDeliveries Gets the recent webhook deliveries limited by `x` records
func GetRecentWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	// Reading the URL parameters
	organizationID := mux.Vars(r)["orgID"]
	webhookID := mux.Vars(r)["webhookID"]

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

	// Validating the given webhook ID for an organisation
	webhook, err := wh.GetByOrgID(webhookID, organizationID)
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}

type webhookDeliveryResp struct {
	ID                      bson.ObjectId       `bson:"_id,omitempty"`
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

// GetWebhookDeliveryByID Gets the payload delivery details for a webhook by ID
func GetWebhookDeliveryByID(w http.ResponseWriter, r *http.Request) {
	// Reading the URL parameters
	organizationID := mux.Vars(r)["orgID"]
	webhookID := mux.Vars(r)["webhookID"]
	deliveryID := mux.Vars(r)["deliveryID"]

	// Validating the given organisation ID
	_, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Validating the given webhook ID for an organisation
	webhook, err := wh.GetByOrgID(webhookID, organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get webhook:%v for organisation: %v", webhookID, organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Get the webhook delivery by ID
	webhookDelivery, err := wh.GetWebhookDeliveryByID(webhook.ID.Hex(), deliveryID)
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// ReDeliverWebhook Redo payload delivery to the webhook
func ReDeliverWebhook(w http.ResponseWriter, r *http.Request) {
	// Reading the URL parameters
	organizationID := mux.Vars(r)["orgID"]
	webhookID := mux.Vars(r)["webhookID"]
	deliveryID := mux.Vars(r)["deliveryID"]

	// Validating the given organisation ID
	_, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Validating the given webhook ID for an organisation
	webhook, err := wh.GetByOrgID(webhookID, organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get webhook:%v for organisation: %v", webhookID, organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Validating the given delivery ID for a webhook
	webhookDelivery, err := wh.GetWebhookDeliveryByID(webhook.ID.Hex(), deliveryID)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

}