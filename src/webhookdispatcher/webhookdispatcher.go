package webhookdispatcher

import (
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
