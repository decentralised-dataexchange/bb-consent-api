package handlerv2

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
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	wh "github.com/bb-consent/api/src/webhooks"
)

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

	organizationID := r.Header.Get(config.OrganizationId)

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
	sanitizedOrgId := common.Sanitize(organizationID)
	sanitizedPayloadURL := common.Sanitize(requestPayload.PayloadURL)

	// Check if webhook with provided payload URL already exists
	count, err := wh.GetWebhookCountByPayloadURL(sanitizedOrgId, sanitizedPayloadURL)
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
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(response)

}
