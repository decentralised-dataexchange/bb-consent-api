package webhook_dispatcher

import (
	"context"
	"strings"

	"github.com/bb-consent/api/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func webhookCollection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("webhooks")
}

func webhookDeliveryCollection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("webhookDeliveries")
}

// GetWebhookByOrgID Gets a webhook by organisation ID and webhook ID
func GetWebhookByOrgID(webhookId, orgID string) (result Webhook, err error) {

	err = webhookCollection().FindOne(context.TODO(), bson.M{"_id": webhookId, "orgid": orgID}).Decode(&result)

	return result, err
}

// AddWebhookDelivery Adds payload delivery details to database for a webhook event
func AddWebhookDelivery(webhookDelivery WebhookDelivery) (WebhookDelivery, error) {

	if len(strings.TrimSpace(webhookDelivery.ID)) < 1 {
		webhookDelivery.ID = primitive.NewObjectID().Hex()
	}

	_, err := webhookDeliveryCollection().InsertOne(context.TODO(), &webhookDelivery)

	return webhookDelivery, err
}

// GetWebhookDeliveryByID Gets payload delivery details by ID
func GetWebhookDeliveryByID(webhookID string, webhookDeliveryId string) (result WebhookDelivery, err error) {

	err = webhookDeliveryCollection().FindOne(context.TODO(), bson.M{"webhookid": webhookID, "_id": webhookDeliveryId}).Decode(&result)

	return result, err
}
