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

// DeleteByUserOrg Deletes the consent by userID, orgID
func DeleteByUserOrg(userID string, orgID string) error {
	s := session()
	defer s.Close()

	return collection(s).Remove(bson.M{"userid": userID, "orgid": orgID})
}

// GetByUserOrg Get all consents of a user in organization
func GetByUserOrg(userID string, orgID string) (Consents, error) {
	s := session()
	defer s.Close()

	var consents Consents
	err := collection(s).Find(bson.M{"userid": userID, "orgid": orgID}).One(&consents)

	return consents, err
}

// Get Get consent by consentID
func Get(consentID string) (Consents, error) {
	s := session()
	defer s.Close()

	var result Consents
	err := collection(s).FindId(bson.ObjectIdHex(consentID)).One(&result)

	return result, err
}

// GetConsentedUsers Get list of users who are consented to an attribute
func GetConsentedUsers(orgID string, purposeID string, attributeID string, startID string, limit int) (userIDs []string, lastID string, err error) {
	s := session()
	defer s.Close()
	c := collection(s)

	limit = 10000
	var results []Consents
	if startID == "" {
		pipeline := []bson.M{
			{"$match": bson.M{"orgid": orgID}},
			{"$unwind": "$purposes"},
			{"$unwind": "$purposes.consents"},
			{"$match": bson.M{
				"purposes.id":                        purposeID,
				"purposes.consents.templateid":       attributeID,
				"purposes.consents.status.consented": bson.M{"$regex": "^A"}},
			},
			{"$limit": limit},
		}
		err = c.Pipe(pipeline).All(&results)
	} else {
		pipeline := []bson.M{
			{"$match": bson.M{"orgid": orgID}},
			{"$unwind": "$purposes"},
			{"$unwind": "$purposes.consents"},
			{"$match": bson.M{
				"purposes.id":                        purposeID,
				"purposes.consents.templateid":       attributeID,
				"purposes.consents.status.consented": bson.M{"$regex": "^A"}},
			},
			{"$limit": limit},
			{"$gt": startID},
		}
		err = c.Pipe(pipeline).All(&results)
	}
	if err != nil {
		return
	}

	for _, item := range results {
		userIDs = append(userIDs, item.UserID)
	}

	if len(results) != 0 && len(results) == (limit) {
		lastID = results[len(results)-1].ID.Hex()
	}

	return
}

// GetPurposeConsentedAllUsers Get all users with at-least one attribute consented in purpose.
func GetPurposeConsentedAllUsers(orgID string, purposeID string, startID string, limit int) (userIDs []string, lastID string, err error) {
	s := session()
	defer s.Close()
	c := collection(s)

	limit = 10000
	var results []Consents
	if startID == "" {
		pipeline := []bson.M{
			{"$match": bson.M{"orgid": orgID}},
			{"$unwind": "$purposes"},
			{"$unwind": "$purposes.consents"},
			{"$match": bson.M{
				"purposes.id":                        purposeID,
				"purposes.consents.status.consented": bson.M{"$regex": "^A"}},
			},
			{"$limit": limit},
		}
		err = c.Pipe(pipeline).All(&results)
	} else {
		pipeline := []bson.M{
			{"$match": bson.M{"orgid": orgID}},
			{"$unwind": "$purposes"},
			{"$unwind": "$purposes.consents"},
			{"$match": bson.M{
				"purposes.id":                        purposeID,
				"purposes.consents.status.consented": bson.M{"$regex": "^A"}},
			},
			{"$limit": limit},
			{"$gt": startID},
		}
		err = c.Pipe(pipeline).All(&results)
	}
	if err != nil {
		return
	}

	keys := make(map[string]bool)
	for _, item := range results {
		if _, value := keys[item.UserID]; !value {
			keys[item.UserID] = true
			userIDs = append(userIDs, item.UserID)
		}
	}

	if len(results) != 0 && len(results) == (limit) {
		lastID = results[len(results)-1].ID.Hex()
	}

	return
}

// UpdatePurposes Update consents purposes
func UpdatePurposes(consents Consents) (Consents, error) {
	s := session()
	defer s.Close()
	c := collection(s)

	return consents, c.Update(bson.M{"_id": consents.ID}, bson.M{"$set": bson.M{"purposes": consents.Purposes}})
}
