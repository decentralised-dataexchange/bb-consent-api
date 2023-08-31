package consenthistory

import (
	"time"

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

// GetByUserID Gets all history of a given userID
func GetByUserID(userID string, startID string, limit int) ([]ConsentHistory, string, error) {
	s := session()
	defer s.Close()

	var results []ConsentHistory
	var err error
	if startID == "" {
		err = collection(s).Find(bson.M{"userid": userID}).Sort("-_id").Limit(limit).All(&results)
	} else {
		err = collection(s).Find(bson.M{"userid": userID, "_id": bson.M{"$lt": bson.ObjectIdHex(startID)}}).Sort("-_id").Limit(limit).All(&results)
	}

	var lastID = ""
	if err == nil {
		if len(results) != 0 && len(results) == (limit) {
			lastID = results[len(results)-1].ID.Hex()
		}
	}

	return results, lastID, err
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

// GetByUserOrgPurposeID Gets all history of a given userID in an organization with purposeID
func GetByUserOrgPurposeID(userID string, orgID string, purposeID string, startID string, limit int) ([]ConsentHistory, string, error) {
	s := session()
	defer s.Close()

	var results []ConsentHistory
	var err error
	if startID == "" {
		err = collection(s).Find(bson.M{"userid": userID, "orgid": orgID, "purposeid": purposeID}).Sort("-_id").Limit(limit).All(&results)
	} else {
		err = collection(s).Find(bson.M{"userid": userID, "orgid": orgID, "purposeid": purposeID, "_id": bson.M{"$lt": bson.ObjectIdHex(startID)}}).Sort("-_id").Limit(limit).All(&results)
	}

	var lastID = ""
	if err == nil {
		if len(results) != 0 && len(results) == (limit) {
			lastID = results[len(results)-1].ID.Hex()
		}
	}

	return results, lastID, err
}

// GetByUserOrgID Gets all history of a given userID in an organization
func GetByUserOrgID(userID string, orgID string, startID string, limit int) ([]ConsentHistory, string, error) {
	s := session()
	defer s.Close()

	var results []ConsentHistory
	var err error
	if startID == "" {
		err = collection(s).Find(bson.M{"userid": userID, "orgid": orgID}).Sort("-_id").Limit(limit).All(&results)
	} else {
		err = collection(s).Find(bson.M{"userid": userID, "orgid": orgID, "_id": bson.M{"$lt": bson.ObjectIdHex(startID)}}).Sort("-_id").Limit(limit).All(&results)
	}

	var lastID = ""
	if err == nil {
		if len(results) != 0 && len(results) == (limit) {
			lastID = results[len(results)-1].ID.Hex()
		}
	}

	return results, lastID, err
}

// GetByDateRange Gets all history of a given userID with date range
func GetByDateRange(userID string, startDate string, endDate string, startID string, limit int) ([]ConsentHistory, string, error) {
	s := session()
	defer s.Close()

	var results []ConsentHistory
	var err error

	//layout := "2006-01-02T15:04:05.00Z"
	sDate, err := time.Parse(time.RFC3339, startDate)

	if err != nil {
		return results, "", err
	}

	eDate, err := time.Parse(time.RFC3339, endDate)

	if err != nil {
		return results, "", err
	}
	sID := bson.NewObjectIdWithTime(sDate)
	eID := bson.NewObjectIdWithTime(eDate)

	if startID == "" {
		err = collection(s).Find(bson.M{"userid": userID, "_id": bson.M{"$gte": sID, "$lt": eID}}).Sort("-_id").Limit(limit).All(&results)
	} else {
		err = collection(s).Find(bson.M{"userid": userID, "_id": bson.M{"$lt": bson.ObjectIdHex(startID), "$gte": sID}}).Sort("-_id").Limit(limit).All(&results)
	}

	var lastID = ""
	if err == nil {
		if len(results) != 0 && len(results) == (limit) {
			lastID = results[len(results)-1].ID.Hex()
		}
	}

	return results, lastID, err
}
