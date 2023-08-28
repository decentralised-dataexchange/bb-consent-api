package consenthistory

import (
	"github.com/bb-consent/api/src/database"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// ConsentHistory HOlds the consent logs
type ConsentHistory struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	UserID    string
	OrgID     string
	PurposeID string
	ConsentID string
	Log       string
}

func session() *mgo.Session {
	return database.DB.Session.Copy()
}

func collection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("consentHistory")
}

// GetLatestByUserOrgPurposeID Gets latest consent history of a given userID in an organization with purposeID
func GetLatestByUserOrgPurposeID(userID string, orgID string, purposeID string) (ConsentHistory, error) {
	s := session()
	defer s.Close()

	var result ConsentHistory
	var err error

	err = collection(s).Find(bson.M{"userid": userID, "orgid": orgID, "purposeid": purposeID}).Sort("-_id").One(&result)

	return result, err
}
