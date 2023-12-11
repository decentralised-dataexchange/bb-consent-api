package webhook

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	wh "github.com/bb-consent/api/internal/webhook"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

func updateWebhookFromAddWebhookRequestBody(requestBody addWebhookReq, newWebhook wh.Webhook) wh.Webhook {
	newWebhook.PayloadURL = requestBody.Webhook.PayloadURL
	newWebhook.ContentType = requestBody.Webhook.ContentType
	newWebhook.SubscribedEvents = requestBody.Webhook.SubscribedEvents
	newWebhook.Disabled = requestBody.Webhook.Disabled
	newWebhook.SecretKey = requestBody.Webhook.SecretKey
	newWebhook.SkipSSLVerification = requestBody.Webhook.SkipSSLVerification
	newWebhook.TimeStamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")

	return newWebhook
}

func validateAddwebhookRequestBody(webhookReq addWebhookReq, organisationId string, webhhokRepo wh.WebhookRepository) error {

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
	count, err := webhhokRepo.GetWebhookCountByPayloadURL(webhookReq.Webhook.PayloadURL)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("webhook with provided payload URL already exists;")
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

type addWebhookReq struct {
	Webhook wh.Webhook `json:"webhook"`
}

type addWebhookResp struct {
	Webhook wh.Webhook `json:"webhook"`
}

// ConfigCreateWebhook
func ConfigCreateWebhook(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Request body
	var webhookReq addWebhookReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &webhookReq)

	// Repository
	webhookRepo := wh.WebhookRepository{}
	webhookRepo.Init(organisationId)

	// Validate request body
	err := validateAddwebhookRequestBody(webhookReq, organisationId, webhookRepo)
	if err != nil {
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var newWebhook wh.Webhook
	newWebhook.ID = primitive.NewObjectID().Hex()
	newWebhook = updateWebhookFromAddWebhookRequestBody(webhookReq, newWebhook)
	newWebhook.OrganisationId = organisationId
	newWebhook.IsDeleted = false

	// Creating webhook
	webhook, err := webhookRepo.CreateWebhook(newWebhook)
	if err != nil {
		m := fmt.Sprintf("Failed to create webhook for organisation:%v", organisationId)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := addWebhookResp{
		Webhook: webhook,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
