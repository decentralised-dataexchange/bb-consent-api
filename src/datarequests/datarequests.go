package datarequests

import (
	"time"

	"github.com/bb-consent/api/src/database"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Data Request type and status const
const (
	DataRequestMaxComments = 14

	DataRequestTypeDelete   = 1
	DataRequestTypeDownload = 2
	DataRequestTypeUpdate   = 3

	DataRequestStatusInitiated              = 1
	DataRequestStatusAcknowledged           = 2
	DataRequestStatusProcessedWithoutAction = 6
	DataRequestStatusProcessedWithAction    = 7
	DataRequestStatusUserCancelled          = 8
)

type iDString struct {
	ID  int
	Str string
}

// Note: Dont change the ID(s) if new type is needed then add at the end

// StatusTypes Array of id and string
var StatusTypes = []iDString{
	iDString{ID: DataRequestStatusInitiated, Str: "Request initiated"},
	iDString{ID: DataRequestStatusAcknowledged, Str: "Request acknowledged"},
	iDString{ID: DataRequestStatusProcessedWithoutAction, Str: "Request processed without action"},
	iDString{ID: DataRequestStatusProcessedWithAction, Str: "Request processed with action"},
	iDString{ID: DataRequestStatusUserCancelled, Str: "Request cancelled by user"},
}

// RequestTypes Array of id and string
var RequestTypes = []iDString{
	iDString{ID: DataRequestTypeDelete, Str: "Delete Data"},
	iDString{ID: DataRequestTypeDownload, Str: "Download Data"},
	iDString{ID: DataRequestTypeUpdate, Str: "Update Data"},
}

// DataRequest Data request information
type DataRequest struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	UserID      string
	OrgID       string
	UserName    string
	ClosedDate  time.Time
	ConsentID   string
	PurposeID   string
	AttributeID string
	Type        int
	State       int
	Comments    [DataRequestMaxComments]string
}

func session() *mgo.Session {
	return database.DB.Session.Copy()
}

func collection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("userDataRequests")
}

// GetStatusTypeStr Get status type string from ID
func GetStatusTypeStr(statusType int) string {
	for _, i := range StatusTypes {
		if i.ID == statusType {
			return i.Str
		}
	}
	return ""
}

// GetRequestTypeStr Get request type string from ID
func GetRequestTypeStr(requestType int) string {
	return RequestTypes[requestType-1].Str
}

// Add Adds access log
func Add(req DataRequest) (DataRequest, error) {
	s := session()
	defer s.Close()

	req.ID = bson.NewObjectId()

	return req, collection(s).Insert(req)
}

// Update Update the req entry
func Update(reqID bson.ObjectId, state int, comments [DataRequestMaxComments]string) (err error) {
	s := session()
	defer s.Close()

	if state >= DataRequestStatusProcessedWithoutAction {
		err = collection(s).Update(bson.M{"_id": reqID}, bson.M{"$set": bson.M{"comments": comments, "state": state, "closeddate": time.Now()}})
	} else {
		err = collection(s).Update(bson.M{"_id": reqID}, bson.M{"$set": bson.M{"comments": comments, "state": state}})
	}
	if err != nil {
		return err
	}
	return nil
}

// GetDataRequestByID Returns the data requests record by ID
func GetDataRequestByID(reqID string) (DataRequest, error) {
	s := session()
	defer s.Close()

	var dataReqest DataRequest
	err := collection(s).FindId(bson.ObjectIdHex(reqID)).One(&dataReqest)

	return dataReqest, err
}

// GetOpenDataRequestsByOrgID Get data requests against orgID
func GetOpenDataRequestsByOrgID(orgID string, startID string, limit int) (results []DataRequest, lastID string, err error) {
	s := session()
	defer s.Close()

	if startID == "" {
		err = collection(s).Find(bson.M{"orgid": orgID, "state": bson.M{"$lt": DataRequestStatusProcessedWithoutAction}}).Sort("-_id").Limit(limit).All(&results)
	} else {
		err = collection(s).Find(bson.M{"orgid": orgID, "state": bson.M{"$lt": DataRequestStatusProcessedWithoutAction},
			"_id": bson.M{"$lt": bson.ObjectIdHex(startID)}}).Sort("-_id").Limit(limit).All(&results)
	}

	lastID = ""
	if err == nil {
		if len(results) != 0 && len(results) == (limit) {
			lastID = results[len(results)-1].ID.Hex()
		}
	}

	return results, lastID, err
}

// GetClosedDataRequestsByOrgID Get data requests against orgID
func GetClosedDataRequestsByOrgID(orgID string, startID string, limit int) (results []DataRequest, lastID string, err error) {
	s := session()
	defer s.Close()

	if startID == "" {
		err = collection(s).Find(bson.M{"orgid": orgID, "state": bson.M{"$gte": DataRequestStatusProcessedWithoutAction}}).Sort("-_id").Limit(limit).All(&results)
	} else {
		err = collection(s).Find(bson.M{"orgid": orgID, "state": bson.M{"$gte": DataRequestStatusProcessedWithoutAction},
			"_id": bson.M{"$lt": bson.ObjectIdHex(startID)}}).Sort("-_id").Limit(limit).All(&results)
	}

	lastID = ""
	if err == nil {
		if len(results) != 0 && len(results) == (limit) {
			lastID = results[len(results)-1].ID.Hex()
		}
	}

	return results, lastID, err
}

// GetOpenDataRequestsByOrgUserID Get data requests against orgID
func GetOpenDataRequestsByOrgUserID(orgID string, userID string, startID string, limit int) (results []DataRequest, lastID string, err error) {
	s := session()
	defer s.Close()

	if startID == "" {
		err = collection(s).Find(bson.M{"orgid": orgID, "userid": userID, "state": bson.M{"$lt": DataRequestStatusProcessedWithoutAction}}).Sort("-_id").Limit(limit).All(&results)
	} else {
		err = collection(s).Find(bson.M{"orgid": orgID, "userid": userID, "state": bson.M{"$lt": DataRequestStatusProcessedWithoutAction},
			"_id": bson.M{"$lt": bson.ObjectIdHex(startID)}}).Sort("-_id").Limit(limit).All(&results)
	}

	lastID = ""
	if err == nil {
		if len(results) != 0 && len(results) == (limit) {
			lastID = results[len(results)-1].ID.Hex()
		}
	}

	return results, lastID, err
}

// GetClosedDataRequestsByOrgUserID Get data requests against orgID
func GetClosedDataRequestsByOrgUserID(orgID string, userID string, startID string, limit int) (results []DataRequest, lastID string, err error) {
	s := session()
	defer s.Close()

	if startID == "" {
		err = collection(s).Find(bson.M{"orgid": orgID, "userid": userID, "state": bson.M{"$gte": DataRequestStatusProcessedWithoutAction}}).Sort("-_id").Limit(limit).All(&results)
	} else {
		err = collection(s).Find(bson.M{"orgid": orgID, "userid": userID, "state": bson.M{"$gte": DataRequestStatusProcessedWithoutAction},
			"_id": bson.M{"$lt": bson.ObjectIdHex(startID)}}).Sort("-_id").Limit(limit).All(&results)
	}

	lastID = ""
	if err == nil {
		if len(results) != 0 && len(results) == (limit) {
			lastID = results[len(results)-1].ID.Hex()
		}
	}

	return results, lastID, err
}

// GetDataRequestsByOrgUserID Get data requests against userID
func GetDataRequestsByOrgUserID(orgID string, userID string, startID string, limit int) (results []DataRequest, lastID string, err error) {
	s := session()
	defer s.Close()

	if startID == "" {
		err = collection(s).Find(bson.M{"orgid": orgID, "userid": userID}).Sort("-_id").Limit(limit).All(&results)
	} else {
		err = collection(s).Find(bson.M{"orgid": orgID, "userid": userID, "_id": bson.M{"$lt": bson.ObjectIdHex(startID)}}).Sort("-_id").Limit(limit).All(&results)
	}

	lastID = ""
	if err == nil {
		if len(results) != 0 && len(results) == (limit) {
			lastID = results[len(results)-1].ID.Hex()
		}
	}

	return results, lastID, err
}

// GetDataRequestsByUserOrgTypeID Get data requests against orgID
func GetDataRequestsByUserOrgTypeID(orgID string, userID string, drType int) (results []DataRequest, err error) {
	s := session()
	defer s.Close()

	err = collection(s).Find(bson.M{"orgid": orgID, "userid": userID, "type": drType}).All(&results)

	return results, err
}