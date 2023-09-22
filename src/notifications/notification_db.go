package notifications

import (
	"context"
	"time"

	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Notification Types
const (
	AttributeConsent = 1
	PurposeChange    = 2
	EulaUpdate       = 3
	Event            = 4
	DataBreach       = 5
)

// Notification data type
type Notification struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Type         int
	Title        string
	UserID       string
	OrgID        string
	ConsentID    string
	PurposeID    string
	ReadStatus   bool
	Timestamp    string
	DataBreachID string
	EventID      string
	AttributeIDs []string
}

func collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("notifications")
}

// Add Adds a notification to the collection
func Add(notification Notification) (Notification, error) {

	notification.ID = primitive.NewObjectID()
	notification.Timestamp = time.Now().Format(time.RFC3339)

	_, err := collection().InsertOne(context.TODO(), &notification)

	return notification, err
}

// GetUnReadCountByUserID gets count of un-read notifications of a given user
func GetUnReadCountByUserID(userID string) (int, error) {

	count, err := collection().CountDocuments(context.TODO(), bson.M{"userid": userID, "readstatus": false})

	return int(count), err
}
