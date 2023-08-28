package misc

import (
	"github.com/bb-consent/api/src/database"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	//DocTypeOrgDataBreach Document type for Data breach
	DocTypeOrgDataBreach = 1

	//DocTypeOrgEvent Document type for Events
	DocTypeOrgEvent = 2
)

// DataBreach stores the Data breach informations
type DataBreach struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
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
	ID      bson.ObjectId `bson:"_id,omitempty"`
	Type    int
	OrgID   string
	Details string
}

func session() *mgo.Session {
	return database.DB.Session.Copy()
}

func collection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("misc")
}

// AddDataBreachNotifications Update the data breach info to organization
func AddDataBreachNotifications(dataBreach DataBreach) error {
	s := session()
	defer s.Close()

	dataBreach.Type = DocTypeOrgDataBreach
	err := collection(s).Insert(dataBreach)
	if err != nil {
		return err
	}
	return nil
}

// AddEventNotifications Update the data breach info to organization
func AddEventNotifications(event Event) error {
	s := session()
	defer s.Close()

	event.Type = DocTypeOrgEvent
	err := collection(s).Insert(event)
	if err != nil {
		return err
	}
	return nil
}
