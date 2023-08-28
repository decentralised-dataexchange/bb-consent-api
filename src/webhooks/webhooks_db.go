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

func session() *mgo.Session {
	return database.DB.Session.Copy()
}

func webhookCollection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("webhooks")
}

// GetActiveWebhooksByOrgID Gets all active webhooks for a particular organisation
func GetActiveWebhooksByOrgID(orgID string) (results []Webhook, err error) {
	s := session()
	defer s.Close()

	err = webhookCollection(s).Find(bson.M{"orgid": orgID, "disabled": false}).All(&results)

	return results, err
}