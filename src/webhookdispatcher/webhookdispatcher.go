package webhookdispatcher

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
	"strconv"
	"strings"
	"time"

	"github.com/bb-consent/api/src/config"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// WebhookEvent Webhook event wrapper
type WebhookEvent struct {
	DeliveryID string      `json:"deliveryID"` // Webhook delivery ID
	WebhookID  string      `json:"webhookID"`  // Webhook endpoint ID
	Timestamp  string      `json:"timestamp"`  // UTC timestamp of webhook triggered data time
	Data       interface{} `json:"data"`       // Event data attribute
	Type       string      `json:"type"`       // Event type for e.g. data.delete.initiated
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

// Webhook Defines the structure for an organisation webhook
type Webhook struct {
	ID                  primitive.ObjectID `bson:"_id,omitempty"` // Webhook ID
	OrgID               string             // Organisation ID
	PayloadURL          string             // Webhook payload URL
	ContentType         string             // Webhook payload content type for e.g application/json
	SubscribedEvents    []string           // Events subscribed for e.g. user.data.delete
	Disabled            bool               // Disabled or not
	SecretKey           string             // For calculating SHA256 HMAC to verify data integrity and authenticity
	SkipSSLVerification bool               // Skip SSL certificate verification or not (expiry is checked)
	TimeStamp           string             // UTC timestamp
}

// WebhookDelivery Details of payload delivery to webhook endpoint
type WebhookDelivery struct {
	ID                      primitive.ObjectID  `bson:"_id,omitempty"` // Webhook delivery ID
	WebhookID               string              // Webhook ID
	UserID                  string              // ID of user who triggered the webhook event
	WebhookEventType        string              // Webhook event type for e.g. data.delete.initiated
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

func WebhookDispatcherInit(webhookConfig *config.Configuration) {

	// Creating a kafka consumer instance
	// https://github.com/edenhill/librdkafka/tree/master/CONFIGURATION.md
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  webhookConfig.Webhooks.KafkaConfig.Broker.URL,
		"group.id":           webhookConfig.Webhooks.KafkaConfig.Broker.GroupID, //
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": true,
	})

	if err != nil {
		panic(err)
	}

	// Subscribing to kafka topics
	err = c.SubscribeTopics([]string{webhookConfig.Webhooks.KafkaConfig.Topic}, nil)
	if err != nil {
		log.Printf("Failed to subscribe to kafka topic:%s; Error: %v", webhookConfig.Webhooks.KafkaConfig.Topic, err)
		panic(err)
	}

	for {
		msg, err := c.ReadMessage(-1)

		if err == nil {

			// Processing webhooks asynchronously
			go func() {

				// For recording execution times
				var executionStartTimeStamp string
				var executionEndTimeStamp string

				// For storing webhook payload delivery details to db
				var webhookDelivery WebhookDelivery

				// Recording webhook processing start timestamp
				executionStartTimeStamp = strconv.FormatInt(time.Now().UTC().Unix(), 10)

				// To store incoming webhook events
				var webhookEvent WebhookEvent

				// To store incoming webhook event type
				var webhookEventType string

				// Unmarshalling the incoming message value bytes to webhook event struct
				err := json.Unmarshal([]byte(msg.Value), &webhookEvent)
				if err != nil {
					log.Printf("Invalid incoming webhook recieved !")
					return
				}

				// Webhook event type
				webhookEventType = string(msg.Key)

				// Webhook event data attribute
				// Converting data attribute to appropriate webhook event struct

				webhookEventData, ok := webhookEvent.Data.(map[string]interface{})
				if !ok {
					log.Printf("Invalid incoming webhook recieved !")
					return
				}

				// Quick fix
				// Retrieving user and organisation ID from webhook data attribute
				userID := webhookEventData["userID"].(string)
				orgID := webhookEventData["organisationID"].(string)

				log.Printf("Processing webhook:%s triggered by user:%s of org:%s for event:%s", webhookEvent.WebhookID, userID, orgID, webhookEventType)

				// Instantiating webhook delivery
				webhookDelivery = WebhookDelivery{
					ID:                      primitive.NewObjectID(),
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
				webhookEvent.DeliveryID = webhookDelivery.ID.Hex()

				// Constructing webhook payload bytes
				requestPayload, _ := json.Marshal(&webhookEvent)

				// Current UTC timestamp
				timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)

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
					executionEndTimeStamp = strconv.FormatInt(time.Now().UTC().Unix(), 10)

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
					executionEndTimeStamp = strconv.FormatInt(time.Now().UTC().Unix(), 10)

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
				executionEndTimeStamp = strconv.FormatInt(time.Now().UTC().Unix(), 10)

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
			}()

		} else {
			// The client will automatically try to recover from all errors.
			fmt.Printf("Webhook dispatcher(Kafka consumer) error: %v (%v)\n", err, msg)
		}
	}

	c.Close()
}
