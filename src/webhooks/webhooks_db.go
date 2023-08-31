package webhooks

import (
	"github.com/bb-consent/api/src/database"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Webhook Defines the structure for an organisation webhook
type Webhook struct {
	ID                  bson.ObjectId `bson:"_id,omitempty"` // Webhook ID
	OrgID               string        // Organisation ID
	PayloadURL          string        // Webhook payload URL
	ContentType         string        // Webhook payload content type for e.g application/json
	SubscribedEvents    []string      // Events subscribed for e.g. user.data.delete
	Disabled            bool          // Disabled or not
	SecretKey           string        // For calculating SHA256 HMAC to verify data integrity and authenticity
	SkipSSLVerification bool          // Skip SSL certificate verification or not (expiry is checked)
	TimeStamp           string        // UTC timestamp
}

// WebhookDelivery Details of payload delivery to webhook endpoint
type WebhookDelivery struct {
	ID                      bson.ObjectId       `bson:"_id,omitempty"` // Webhook delivery ID
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

func session() *mgo.Session {
	return database.DB.Session.Copy()
}

func webhookCollection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("webhooks")
}

func webhookDeliveryCollection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("webhookDeliveries")
}

// CreateWebhook Adds a webhook for an organisation
func CreateWebhook(webhook Webhook) (Webhook, error) {
	s := session()
	defer s.Close()

	webhook.ID = bson.NewObjectId()

	return webhook, webhookCollection(s).Insert(&webhook)
}

// GetByOrgID Gets a webhook by organisation ID and webhook ID
func GetByOrgID(webhookID, orgID string) (result Webhook, err error) {
	s := session()
	defer s.Close()

	err = webhookCollection(s).Find(bson.M{"_id": bson.ObjectIdHex(webhookID), "orgid": orgID}).One(&result)

	return result, err
}

// DeleteWebhook Deletes a webhook for an organisation
func DeleteWebhook(webhookID string) error {
	s := session()
	defer s.Close()

	return webhookCollection(s).RemoveId(bson.ObjectIdHex(webhookID))
}

// UpdateWebhook Updates a webhook for an organization
func UpdateWebhook(webhook Webhook) (Webhook, error) {
	s := session()
	defer s.Close()

	err := webhookCollection(s).UpdateId(webhook.ID, webhook)
	return webhook, err
}

// GetActiveWebhooksByOrgID Gets all active webhooks for a particular organisation
func GetActiveWebhooksByOrgID(orgID string) (results []Webhook, err error) {
	s := session()
	defer s.Close()

	err = webhookCollection(s).Find(bson.M{"orgid": orgID, "disabled": false}).All(&results)

	return results, err
}

// GetWebhookCountByPayloadURL Gets the count of webhooks with same payload URL for an organisation
func GetWebhookCountByPayloadURL(orgID string, payloadURL string) (count int, err error) {
	s := session()
	defer s.Close()

	count, err = webhookCollection(s).Find(bson.M{"orgid": orgID, "payloadurl": payloadURL}).Count()

	return count, err
}

// GetAllWebhooksByOrgID Gets all webhooks for a given organisation
func GetAllWebhooksByOrgID(orgID string) (results []Webhook, err error) {
	s := session()
	defer s.Close()

	err = webhookCollection(s).Find(bson.M{"orgid": orgID}).Sort("-timestamp").All(&results)

	return results, err
}

// GetLastWebhookDelivery Gets the last delivery for a webhook
func GetLastWebhookDelivery(webhookID string) (result WebhookDelivery, err error) {
	s := session()
	defer s.Close()

	err = webhookDeliveryCollection(s).Find(bson.M{"webhookid": webhookID}).Sort("-executionstarttimestamp").One(&result)

	return result, err
}

// GetWebhookByPayloadURL Get the webhook for an organisation by payload URL
func GetWebhookByPayloadURL(orgID string, payloadURL string) (result Webhook, err error) {
	s := session()
	defer s.Close()

	err = webhookCollection(s).Find(bson.M{"orgid": orgID, "payloadurl": payloadURL}).One(&result)

	return result, err
}

// GetWebhookDeliveryByID Gets payload delivery details by ID
func GetWebhookDeliveryByID(webhookID string, webhookDeliveryID string) (result WebhookDelivery, err error) {
	s := session()
	defer s.Close()

	err = webhookDeliveryCollection(s).Find(bson.M{"webhookid": webhookID, "_id": bson.ObjectIdHex(webhookDeliveryID)}).One(&result)

	return result, err
}

// GetAllDeliveryByWebhookID Gets all webhook deliveries for a webhook
func GetAllDeliveryByWebhookID(webhookID string, startID string, limit int) (results []WebhookDelivery, lastID string, err error) {
	s := session()
	defer s.Close()

	if startID == "" {
		err = webhookDeliveryCollection(s).Find(bson.M{"webhookid": webhookID}).Sort("-executionstarttimestamp").Limit(limit).All(&results)
	} else {
		err = webhookDeliveryCollection(s).Find(bson.M{"webhookid": webhookID, "_id": bson.M{"$lt": bson.ObjectIdHex(startID)}}).Sort("-executionstarttimestamp").Limit(limit).All(&results)
	}

	lastID = ""
	if err == nil {
		if len(results) != 0 && len(results) == (limit) {
			lastID = results[len(results)-1].ID.Hex()
		}
	}

	return results, lastID, err
}
