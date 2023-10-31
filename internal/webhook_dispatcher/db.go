package webhook_dispatcher

import (
	"context"

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
func GetWebhookByOrgID(webhookID, orgID string) (result Webhook, err error) {
	webhookId, err := primitive.ObjectIDFromHex(webhookID)
	if err != nil {
		return result, err
	}

	err = webhookCollection().FindOne(context.TODO(), bson.M{"_id": webhookId, "orgid": orgID}).Decode(&result)

	return result, err
}

// AddWebhookDelivery Adds payload delivery details to database for a webhook event
func AddWebhookDelivery(webhookDelivery WebhookDelivery) (WebhookDelivery, error) {

	if webhookDelivery.ID == primitive.NilObjectID {
		webhookDelivery.ID = primitive.NewObjectID()
	}

	_, err := webhookDeliveryCollection().InsertOne(context.TODO(), &webhookDelivery)

	return webhookDelivery, err
}
