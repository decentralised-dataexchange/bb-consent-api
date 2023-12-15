package webhook_dispatcher

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebhookEvent Webhook event wrapper
type WebhookEvent struct {
	DeliveryID string                    `json:"deliveryID"` // Webhook delivery ID
	WebhookID  string                    `json:"webhookID"`  // Webhook endpoint ID
	Timestamp  string                    `json:"timestamp"`  // UTC timestamp of webhook triggered data time
	Data       ConsentRecordWebhookEvent `json:"data"`       // Event data attribute
	Type       string                    `json:"type"`       // Event type for e.g. data.delete.initiated
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

type Webhook struct {
	ID                  string   `json:"id" bson:"_id,omitempty"`           // Webhook ID
	OrganisationId      string   `json:"orgId" bson:"orgid"`                // Organisation ID
	PayloadURL          string   `json:"payloadUrl" valid:"required"`       // Webhook payload URL
	ContentType         string   `json:"contentType" valid:"required"`      // Webhook payload content type for e.g application/json
	SubscribedEvents    []string `json:"subscribedEvents" valid:"required"` // Events subscribed for e.g. user.data.delete
	Disabled            bool     `json:"disabled"`                          // Disabled or not
	SecretKey           string   `json:"secretKey" valid:"required"`        // For calculating SHA256 HMAC to verify data integrity and authenticity
	SkipSSLVerification bool     `json:"skipSslVerification"`               // Skip SSL certificate verification or not (expiry is checked)
	TimeStamp           string   `json:"timestamp" valid:"required"`        // UTC timestamp
	IsDeleted           bool     `json:"-"`
}

// WebhookDelivery Details of payload delivery to webhook endpoint
type WebhookDelivery struct {
	ID                      string              `bson:"_id,omitempty"` // Webhook delivery ID
	WebhookID               string              // Webhook ID
	UserID                  string              // ID of user who triggered the webhook event
	WebhookEventType        string              // Webhook event type for e.g. data.delete.initiated
	RequestHeaders          map[string][]string // HTTP headers posted to webhook endpoint
	RequestPayload          WebhookEvent        // JSON payload posted to webhook endpoint
	ResponseHeaders         map[string][]string // HTTP response headers received from webhook endpoint
	ResponseBody            string              // HTTP response body received from webhook endpoint in bytes
	ResponseStatusCode      int                 // HTTP response status code
	ResponseStatusStr       string              // HTTP response status string
	ExecutionStartTimeStamp string              // UTC timestamp when webhook execution started
	ExecutionEndTimeStamp   string              // UTC timestamp when webhook execution ended
	Status                  string              // Status of webhook delivery for e.g. failed or completed
	StatusDescription       string              // Describe the status for e.g. Reason for failure
}

func ProcessWebhooks(webhookEventType string, value []byte) {
	// For recording execution times
	var executionStartTimeStamp string
	var executionEndTimeStamp string

	// For storing webhook payload delivery details to db
	var webhookDelivery WebhookDelivery

	// Recording webhook processing start timestamp
	executionStartTimeStamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")

	// To store incoming webhook events
	var webhookEvent WebhookEvent

	// Unmarshalling the incoming message value bytes to webhook event struct
	err := json.Unmarshal([]byte(value), &webhookEvent)
	if err != nil {
		log.Printf("Invalid incoming webhook recieved !")
		return
	}

	// Webhook event data attribute
	// Converting data attribute to appropriate webhook event struct

	webhookEventData := webhookEvent.Data

	// Quick fix
	// Retrieving user and organisation ID from webhook data attribute
	userID := webhookEventData.IndividualId
	orgID := webhookEventData.OrganisationId

	log.Printf("Processing webhook:%s triggered by user:%s of org:%s for event:%s", webhookEvent.WebhookID, userID, orgID, webhookEventType)

	// Instantiating webhook delivery
	webhookDelivery = WebhookDelivery{
		ID:                      primitive.NewObjectID().Hex(),
		WebhookID:               webhookEvent.WebhookID,
		UserID:                  userID,
		WebhookEventType:        webhookEventType,
		ExecutionStartTimeStamp: executionStartTimeStamp,
	}

	// Fetch webhook by ID
	webhook, err := GetWebhookByOrgID(webhookEvent.WebhookID, orgID)
	if err != nil {
		log.Printf("Failed to fetch by webhook from db;Failed processing webhook:%s triggered by user:%s of org:%s for event:%s", webhookEvent.WebhookID, userID, orgID, webhookEventType)
		return
	}

	// Checking if the webhook is disabled or not
	if webhook.Disabled {
		log.Printf("Webhook is disabled;Failed processing webhook:%s triggered by user:%s of org:%s for event:%s", webhookEvent.WebhookID, userID, orgID, webhookEventType)
		return
	}

	// Getting the webhook secret key
	secretKey := webhook.SecretKey

	// Updating webhook event payload with delivery ID
	webhookEvent.DeliveryID = webhookDelivery.ID

	// Constructing webhook payload bytes
	requestPayload, _ := json.Marshal(&webhookEvent)

	// Current UTC timestamp
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	// Constructing SHA256 payload
	sha256Payload := timestamp + "." + string(requestPayload)

	// Create a new HMAC by defining the hash type and the key (as byte array)
	h := hmac.New(sha256.New, []byte(secretKey))

	// Write Data to it
	h.Write([]byte(sha256Payload))

	// Get result and encode as hexadecimal string
	sha := hex.EncodeToString(h.Sum(nil))

	// Constructing HTTP request instance based payload content type
	var req *http.Request
	if webhook.ContentType == PayloadContentTypes[PayloadContentTypeFormURLEncoded] {
		// x-www-form-urlencoded payload
		data := url.Values{}
		data.Set("payload", string(requestPayload))

		req, _ = http.NewRequest("POST", webhook.PayloadURL, strings.NewReader(data.Encode()))
	} else {
		req, _ = http.NewRequest("POST", webhook.PayloadURL, bytes.NewBuffer(requestPayload))
	}

	// Adding HTTP headers
	// If secret key is defined, then add X-IGrant-Signature header for checking data integrity and authenticity
	if len(strings.TrimSpace(secretKey)) > 0 {
		req.Header.Set("X-IGrant-Signature", fmt.Sprintf("t=%s,sig=%s", timestamp, sha))
	}

	req.Header.Set("Content-Type", webhook.ContentType)
	req.Header.Set("User-Agent", "IGrant-Hookshot/1.0")
	req.Header.Set("Accept", "*/*")

	// Skip SSL certificate verification or not
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: webhook.SkipSSLVerification},
	}

	client := &http.Client{Transport: transCfg}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("HTTP POST request failed err:%v;Failed processing webhook:%s triggered by user:%s of org:%s for event:%s", err.Error(), webhookEvent.WebhookID, userID, orgID, webhookEventType)

		// Recording webhook processing end timestamp
		executionEndTimeStamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")

		// Recording webhook delivery details to db
		webhookDelivery.RequestHeaders = req.Header
		webhookDelivery.RequestPayload = webhookEvent
		webhookDelivery.StatusDescription = fmt.Sprintf("Error performing HTTP POST for the webhook endpoint:%s", webhook.PayloadURL)
		webhookDelivery.Status = DeliveryStatus[DeliveryStatusFailed]
		webhookDelivery.ExecutionEndTimeStamp = executionEndTimeStamp

		_, err = AddWebhookDelivery(webhookDelivery)
		if err != nil {
			log.Printf("Failed to save webhook delivery details to db;Failed processing webhook:%s triggered by user:%s of org:%s for event:%s", webhookEvent.WebhookID, userID, orgID, webhookEventType)
			return
		}

		return
	}
	defer resp.Body.Close()

	respBodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {

		// Recording webhook processing end timestamp
		executionEndTimeStamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")

		// Recording webhook delivery details to db
		webhookDelivery.RequestHeaders = req.Header
		webhookDelivery.RequestPayload = webhookEvent
		webhookDelivery.ResponseHeaders = resp.Header
		webhookDelivery.ResponseStatusCode = resp.StatusCode
		webhookDelivery.ResponseStatusStr = resp.Status
		webhookDelivery.ExecutionEndTimeStamp = executionEndTimeStamp
		webhookDelivery.Status = DeliveryStatus[DeliveryStatusCompleted]

		_, err = AddWebhookDelivery(webhookDelivery)
		if err != nil {
			log.Printf("Failed to save webhook delivery details to db;Failed processing webhook:%s triggered by user:%s of org:%s for event:%s", webhookEvent.WebhookID, userID, orgID, webhookEventType)
			return
		}

		log.Printf("Failed to read webhook endpoint response;Failed processing webhook:%s triggered by user:%s of org:%s for event:%s", webhookEvent.WebhookID, userID, orgID, webhookEventType)
		return
	}

	// Recording webhook processing end timestamp
	executionEndTimeStamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")

	// Recording webhook delivery details to db
	webhookDelivery.RequestHeaders = req.Header
	webhookDelivery.RequestPayload = webhookEvent
	webhookDelivery.ResponseHeaders = resp.Header
	webhookDelivery.ResponseBody = string(respBodyBytes)
	webhookDelivery.ResponseStatusCode = resp.StatusCode
	webhookDelivery.ResponseStatusStr = resp.Status
	webhookDelivery.ExecutionEndTimeStamp = executionEndTimeStamp
	webhookDelivery.Status = DeliveryStatus[DeliveryStatusCompleted]

	_, err = AddWebhookDelivery(webhookDelivery)
	if err != nil {
		log.Printf("Failed to save webhook delivery details to db;Failed processing webhook:%s triggered by user:%s of org:%s for event:%s", webhookEvent.WebhookID, userID, orgID, webhookEventType)
		return
	}
}
