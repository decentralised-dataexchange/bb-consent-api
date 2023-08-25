package notifications

import (
	"time"

	"github.com/bb-consent/api/src/database"
	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
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
	ID           bson.ObjectId `bson:"_id,omitempty"`
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

func session() *mgo.Session {
	return database.DB.Session.Copy()
}

func collection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("notifications")
}

// Add Adds a notification to the collection
func Add(notification Notification) (Notification, error) {
	s := session()
	defer s.Close()

	notification.ID = bson.NewObjectId()
	notification.Timestamp = time.Now().Format(time.RFC3339)

	return notification, collection(s).Insert(&notification)
}

// GetUnReadCountByUserID gets count of un-read notifications of a given user
func GetUnReadCountByUserID(userID string) (count int, err error) {
	s := session()
	defer s.Close()

	count, err = collection(s).Find(bson.M{"userid": userID, "readstatus": false}).Count()

	return count, err
}
