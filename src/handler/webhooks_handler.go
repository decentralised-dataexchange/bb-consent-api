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
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/org"
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
