package webhook

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	wh "github.com/bb-consent/api/internal/webhook"
	"github.com/gorilla/mux"
)

func validateUpdatewebhookRequestBody(webhookReq updateWebhookReq, currentWebhook wh.Webhook, organisationId string, webhookId string, webhhokRepo wh.WebhookRepository) error {

	// Validating request payload struct
	_, err := govalidator.ValidateStruct(webhookReq)
	if err != nil {
		return err
	}
	// Check if the webhook endpoint contains http:// or https://
	if !(strings.HasPrefix(webhookReq.Webhook.PayloadURL, "https://") || strings.HasPrefix(webhookReq.Webhook.PayloadURL, "http://")) {
		return errors.New("please prefix the endpoint URL with https:// or http://;")
	}

	// Check if webhook with provided payload URL already exists
	tempWebhook, err := webhhokRepo.GetWebhookByPayloadURL(webhookReq.Webhook.PayloadURL)
	if err == nil {
		if tempWebhook.ID != webhookId {
			return errors.New("webhook with provided payload URL already exists")
		}
	}

	// Check if subscribed event type(s) array is empty
	if len(webhookReq.Webhook.SubscribedEvents) == 0 {
		return errors.New("provide atleast 1 event type;")
	}

	// Check if subscribed event type(s) contains duplicates
	webhookReq.Webhook.SubscribedEvents = uniqueSlice(webhookReq.Webhook.SubscribedEvents)

	// Check the subscribed event type ID(s) provided is valid
	var isValidSubscribedEvents bool
	for _, subscribedEventType := range webhookReq.Webhook.SubscribedEvents {
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
		return errors.New("please provide a valid event type")
	}

	// Check if the content type ID provided is valid
	isValidContentType := false
	for _, payloadContentType := range wh.PayloadContentTypes {
		if webhookReq.Webhook.ContentType == payloadContentType {
			isValidContentType = true
		}
	}

	if !isValidContentType {
		return errors.New("please provide a valid content type")
	}
	return nil
}

func updateWebhookFromUpdateWebhookRequestBody(requestBody updateWebhookReq, toBeUpdatedWebhook wh.Webhook) wh.Webhook {
	toBeUpdatedWebhook.PayloadURL = requestBody.Webhook.PayloadURL
	toBeUpdatedWebhook.ContentType = requestBody.Webhook.ContentType
	toBeUpdatedWebhook.SubscribedEvents = requestBody.Webhook.SubscribedEvents
	toBeUpdatedWebhook.Disabled = requestBody.Webhook.Disabled
	toBeUpdatedWebhook.SecretKey = requestBody.Webhook.SecretKey
	toBeUpdatedWebhook.SkipSSLVerification = requestBody.Webhook.SkipSSLVerification

	return toBeUpdatedWebhook
}

type updateWebhookReq struct {
	Webhook wh.Webhook `json:"webhook"`
}

type updateWebhookResp struct {
	Webhook wh.Webhook `json:"webhook"`
}

// ConfigUpdateWebhook
func ConfigUpdateWebhook(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	webhookId := mux.Vars(r)[config.WebhookId]
	webhookId = common.Sanitize(webhookId)

	// Request body
	var webhookReq updateWebhookReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &webhookReq)

	// Repository
	webhookRepo := wh.WebhookRepository{}
	webhookRepo.Init(organisationId)

	// Fetching webhook by ID
	toBeUpdatedWebhook, err := webhookRepo.GetByOrgID(webhookId)
	if err != nil {
		m := fmt.Sprintf("Failed to get webhook:%v for organisation: %v", webhookId, organisationId)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	// Validate request body
	err = validateUpdatewebhookRequestBody(webhookReq, toBeUpdatedWebhook, organisationId, webhookId, webhookRepo)
	if err != nil {
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	toBeUpdatedWebhook = updateWebhookFromUpdateWebhookRequestBody(webhookReq, toBeUpdatedWebhook)

	// Save to db
	savedWebhook, err := webhookRepo.UpdateWebhook(toBeUpdatedWebhook)
	if err != nil {
		m := fmt.Sprintf("Failed to update webhook:%v for organisation: %v", webhookId, organisationId)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := updateWebhookResp{
		Webhook: savedWebhook,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
