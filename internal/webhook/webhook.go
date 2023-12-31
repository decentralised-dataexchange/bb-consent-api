package webhook

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/bb-consent/api/internal/actionlog"
	"github.com/bb-consent/api/internal/config"
	daRecord "github.com/bb-consent/api/internal/dataagreement_record"
	"github.com/bb-consent/api/internal/individual"
	"github.com/bb-consent/api/internal/webhook_dispatcher"
)

// Event type const
const (
	// User request events
	EventTypeDataDeleteInitiated   = 10
	EventTypeDataDownloadInitiated = 11
	EventTypeDataUpdateInitiated   = 12
	EventTypeDataDeleteCancelled   = 13
	EventTypeDataDownloadCancelled = 14
	EventTypeDataUpdateCancelled   = 15

	// Consent events
	EventTypeConsentAllowed    = 30
	EventTypeConsentDisAllowed = 31
	EventTypeConsentAutoExpiry = 32

	// Organisation subscription events
	EventTypeOrgSubscribed   = 50
	EventTypeOrgUnSubscribed = 51
)

// EventTypes Map of webhook event type id and name
var EventTypes = map[int]string{
	EventTypeDataDeleteInitiated:   "data.delete.initiated",
	EventTypeDataDownloadInitiated: "data.download.initiated",
	EventTypeDataUpdateInitiated:   "data.update.initiated",
	EventTypeDataDeleteCancelled:   "data.delete.cancelled",
	EventTypeDataDownloadCancelled: "data.download.cancelled",
	EventTypeDataUpdateCancelled:   "data.update.cancelled",
	EventTypeConsentAllowed:        "consent.allowed",
	EventTypeConsentDisAllowed:     "consent.disallowed",
	EventTypeConsentAutoExpiry:     "consent.auto_expiry",
	EventTypeOrgSubscribed:         "org.subscribed",
	EventTypeOrgUnSubscribed:       "org.unsubscribed",
}

// WebhooksConfiguration Stores webhooks configuration
var WebhooksConfiguration config.WebhooksConfig

// Init Initializes webhooks configuration
func Init(config *config.Configuration) {
	WebhooksConfiguration = config.Webhooks
}

// OrgSubscriptionWebhookEvent Details of organisation subscription event
type OrgSubscriptionWebhookEvent struct {
	OrganisationID string `json:"organisationID"`
	UserID         string `json:"userID"`
}

// GetOrganisationID Returns organisation ID
func (e OrgSubscriptionWebhookEvent) GetOrganisationID() string {
	return e.OrganisationID
}

// GetUserID Returns user ID
func (e OrgSubscriptionWebhookEvent) GetUserID() string {
	return e.UserID
}

// WebhookEventData Interface defining the functions a webhook event data struct must implement
type WebhookEventData interface {
	GetOrganisationID() string
	GetUserID() string
}

// WebhookEvent Defines the structure for webhook event
type WebhookEvent struct {
	WebhookID string      `json:"webhookID"` // Webhook endpoint ID
	Timestamp string      `json:"timestamp"` // UTC timestamp of webhook triggered data time
	Data      interface{} `json:"data"`      // Event data attribute
	Type      string      `json:"type"`      // Event type for e.g. data.delete.initiated
}

// Payload content type const
const (
	// Payload will be posted as json body
	PayloadContentTypeJSON = 112

	// Payload will be stringified and posted as form under `payload` key
	PayloadContentTypeFormURLEncoded = 113
)

// PayloadContentTypes Available data format for payload to be posted to webhook
var PayloadContentTypes = map[int]string{
	PayloadContentTypeJSON:           "application/json",
	PayloadContentTypeFormURLEncoded: "application/x-www-form-urlencoded",
}

// Delivery status const
const (
	DeliveryStatusCompleted = 212
	DeliveryStatusFailed    = 213
)

// DeliveryStatus Indicating the payload delivery status to webhook
var DeliveryStatus = map[int]string{
	DeliveryStatusCompleted: "completed",
	DeliveryStatusFailed:    "failed",
}

type ConsentRecordWebhookEvent struct {
	ConsentRecordId           string `json:"consentRecordId"`
	DataAgreementId           string `json:"dataAgreementId"`
	DataAgreementRevisionId   string `json:"dataAgreementRevisionId"`
	DataAgreementRevisionHash string `json:"dataAgreementRevisionHash"`
	IndividualId              string `json:"individualId"`
	OptIn                     bool   `json:"optIn"`
	State                     string `json:"state"`
	SignatureId               string `json:"signatureId"`
	OrganisationId            string `json:"organisationId"`
}

// GetOrganisationID Returns organisation ID
func (e ConsentRecordWebhookEvent) GetOrganisationID() string {
	return e.OrganisationId
}

// GetUserID Returns user ID
func (e ConsentRecordWebhookEvent) GetUserID() string {
	return e.IndividualId
}

// PingWebhook Pings webhook payload URL to check the status
func PingWebhook(webhook Webhook) (req *http.Request, resp *http.Response, executionStartTimeStamp string, executionEndTimeStamp string, err error) {
	executionStartTimeStamp = strconv.FormatInt(time.Now().UTC().Unix(), 10)

	// Initializing a http request object and configuring necessary HTTP headers
	req, _ = http.NewRequest("POST", webhook.PayloadURL, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "IGrant-Hookshot/1.0")
	req.Header.Set("Accept", "*/*")

	// Defining a custom HTTP transport to control SSL certificate verification
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: webhook.SkipSSLVerification},
	}

	client := &http.Client{Transport: transCfg}
	resp, err = client.Do(req)

	executionEndTimeStamp = strconv.FormatInt(time.Now().UTC().Unix(), 10)

	return req, resp, executionStartTimeStamp, executionEndTimeStamp, err
}

// TriggerWebhooks Trigger webhooks based on event type
func TriggerWebhooks(webhookEventData WebhookEventData, webhookEventType string) {

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(webhookEventData.GetOrganisationID())

	// Get the user
	individual, err := individualRepo.Get(webhookEventData.GetUserID())
	if err != nil {
		log.Printf("Failed to fetch user details;Failed to trigger webhook for event:<%s>, org:<%s>", webhookEventType, webhookEventData.GetOrganisationID())
		return
	}

	// Get the active webhooks for the organisation
	activeWebhooks, err := GetActiveWebhooksByOrgID(webhookEventData.GetOrganisationID())
	if err != nil {
		log.Printf("Failed to fetch active webhooks;Failed to trigger webhook for event:<%s>, user:<%s>, org:<%s>", webhookEventType, individual.Id, webhookEventData.GetOrganisationID())
		return
	}

	// Filtering the webhooks that are subscribed to the event
	var toBeProcessedWebhooks []Webhook
	for _, activeWebhook := range activeWebhooks {
		for _, subscribedEvent := range activeWebhook.SubscribedEvents {
			if subscribedEvent == webhookEventType {
				toBeProcessedWebhooks = append(toBeProcessedWebhooks, activeWebhook)
				break
			}
		}
	}

	for _, toBeProcessedWebhook := range toBeProcessedWebhooks {
		// Constructing webhook payload
		we := WebhookEvent{
			WebhookID: toBeProcessedWebhook.ID,
			Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05Z"),
			Data:      webhookEventData,
			Type:      webhookEventType,
		}

		// Converting the webhook event data to bytes
		b, err := json.Marshal(we)
		if err != nil {
			log.Printf("Failed to convert webhook event data to bytes, error:%v, Failed to trigger webhook for event:<%s>, user:<%s>, org:<%s>", err.Error(), webhookEventType, individual.Id, webhookEventData.GetOrganisationID())
			return
		}

		go webhook_dispatcher.ProcessWebhooks(webhookEventType, b)

		// Log webhook calls in webhooks category
		aLog := fmt.Sprintf("Organization webhook: %v triggered by user: %v by event: %v", toBeProcessedWebhook.PayloadURL, individual.Email, webhookEventType)
		actionlog.LogOrgWebhookCalls(individual.Id, individual.Email, webhookEventData.GetOrganisationID(), aLog)
	}

}

// TriggerOrgSubscriptionWebhookEvent Trigger webhook for organisation subscription related events
func TriggerOrgSubscriptionWebhookEvent(userID, orgID string, eventType string) {

	// Constructing webhook event data attribute
	orgSubscriptionWebhookEvent := OrgSubscriptionWebhookEvent{
		OrganisationID: orgID,
		UserID:         userID,
	}

	// triggering the webhook
	TriggerWebhooks(orgSubscriptionWebhookEvent, eventType)
}

// TriggerConsentWebhookEvent Trigger webhook for consent related events
func TriggerConsentWebhookEvent(consentRecord daRecord.DataAgreementRecord, organisationId string, eventType string) {

	// Constructing webhook event data attribute
	consentRecordWebhookEvent := ConsentRecordWebhookEvent{
		ConsentRecordId:           consentRecord.Id,
		DataAgreementId:           consentRecord.DataAgreementId,
		DataAgreementRevisionId:   consentRecord.DataAgreementRevisionId,
		DataAgreementRevisionHash: consentRecord.DataAgreementRevisionHash,
		IndividualId:              consentRecord.IndividualId,
		OptIn:                     consentRecord.OptIn,
		State:                     consentRecord.State,
		SignatureId:               consentRecord.SignatureId,
		OrganisationId:            consentRecord.OrganisationId,
	}

	for _, e := range WebhooksConfiguration.Events {
		if e == eventType {
			// triggering the webhook
			TriggerWebhooks(consentRecordWebhookEvent, eventType)
		}

	}
}
