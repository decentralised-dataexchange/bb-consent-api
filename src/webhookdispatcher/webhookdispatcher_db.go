package webhookdispatcher

import (
	"github.com/bb-consent/api/src/database"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func session() *mgo.Session {
	return database.DB.Session.Copy()
}

func webhookCollection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("webhooks")
}

func webhookDeliveryCollection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("webhookDeliveries")
}

// GetWebhookByOrgID Gets a webhook by organisation ID and webhook ID
func GetWebhookByOrgID(webhookID, orgID string) (result Webhook, err error) {
	s := session()
	defer s.Close()

	err = webhookCollection(s).Find(bson.M{"_id": bson.ObjectIdHex(webhookID), "orgid": orgID}).One(&result)

	return result, err
}

// AddWebhookDelivery Adds payload delivery details to database for a webhook event
func AddWebhookDelivery(webhookDelivery WebhookDelivery) (WebhookDelivery, error) {
	s := session()
	defer s.Close()

	if webhookDelivery.ID == "" {
		webhookDelivery.ID = bson.NewObjectId()
	}

	return webhookDelivery, webhookDeliveryCollection(s).Insert(&webhookDelivery)
}
