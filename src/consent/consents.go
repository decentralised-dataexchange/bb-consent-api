package consent

import (
	"time"

	"github.com/bb-consent/api/src/database"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type consentStatus struct {
	Consented string
	TimeStamp time.Time
	Days      int
}

// Consent data type
type Consent struct {
	Status     consentStatus
	Value      string //Description??
	TemplateID string
}

// Purpose data type
type Purpose struct {
	ID       string `bson:"id,omitempty"`
	AllowAll bool
	Consents []Consent
}

// Consents data type
type Consents struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	OrgID    string
	UserID   string
	Purposes []Purpose
}

func session() *mgo.Session {
	return database.DB.Session.Copy()
}

func collection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("consents")
}

// Add Adds an consent to the collection
func Add(consent Consents) (Consents, error) {
	s := session()
	defer s.Close()

	consent.ID = bson.NewObjectId()
	return consent, collection(s).Insert(&consent)
}
