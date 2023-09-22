package misc

import (
	"context"

	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	//DocTypeOrgDataBreach Document type for Data breach
	DocTypeOrgDataBreach = 1

	//DocTypeOrgEvent Document type for Events
	DocTypeOrgEvent = 2
)

// DataBreach stores the Data breach informations
type DataBreach struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Type        int
	OrgID       string
	HeadLine    string
	UsersCount  int
	DpoEmail    string
	Consequence string
	Measures    string
}

// Event stores event related information.
type Event struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Type    int
	OrgID   string
	Details string
}

func collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("misc")
}

// AddDataBreachNotifications Update the data breach info to organization
func AddDataBreachNotifications(dataBreach DataBreach) error {

	dataBreach.Type = DocTypeOrgDataBreach
	_, err := collection().InsertOne(context.TODO(), dataBreach)
	if err != nil {
		return err
	}
	return nil
}

// AddEventNotifications Update the data breach info to organization
func AddEventNotifications(event Event) error {

	event.Type = DocTypeOrgEvent
	_, err := collection().InsertOne(context.TODO(), event)
	if err != nil {
		return err
	}
	return nil
}
