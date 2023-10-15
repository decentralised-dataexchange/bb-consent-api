package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	wh "github.com/bb-consent/api/src/webhooks"
	"github.com/gorilla/mux"
)

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
	organizationID := r.Header.Get(config.OrganizationId)
	webhookID := mux.Vars(r)[config.WebhookId]

	fmt.Printf("Mux vars : %v\n", mux.Vars(r))

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

	sanitizedPayloadURL := common.Sanitize(requestPayload.PayloadURL)

	// Check if webhook with provided payload URL already exists
	tempWebhook, err := wh.GetWebhookByPayloadURL(sanitizedOrgId, sanitizedPayloadURL)
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
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
