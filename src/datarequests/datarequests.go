package datarequests

import (
	"context"
	"time"

	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	ID          primitive.ObjectID `bson:"_id,omitempty"`
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

func collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("userDataRequests")
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

	req.ID = primitive.NewObjectID()

	_, err := collection().InsertOne(context.TODO(), req)

	return req, err
}

// Update Update the req entry
func Update(reqID primitive.ObjectID, state int, comments [DataRequestMaxComments]string) (err error) {

	if state >= DataRequestStatusProcessedWithoutAction {
		_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": reqID}, bson.M{"$set": bson.M{"comments": comments, "state": state, "closeddate": time.Now()}})
	} else {
		_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": reqID}, bson.M{"$set": bson.M{"comments": comments, "state": state}})
	}
	if err != nil {
		return err
	}
	return nil
}

// GetDataRequestByID Returns the data requests record by ID
func GetDataRequestByID(reqID string) (DataRequest, error) {
	var dataReqest DataRequest

	reqId, err := primitive.ObjectIDFromHex(reqID)
	if err != nil {
		return dataReqest, err
	}

	err = collection().FindOne(context.TODO(), bson.M{"_id": reqId}).Decode(&dataReqest)

	return dataReqest, err
}

// GetOpenDataRequestsByOrgID Get data requests against orgID
func GetOpenDataRequestsByOrgID(orgID string, startID string, limit int) (results []DataRequest, lastID string, err error) {

	filter := bson.M{
		"orgid": orgID,
		"state": bson.M{"$lt": DataRequestStatusProcessedWithoutAction},
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})
	findOptions.SetLimit(int64(limit))

	if startID != "" {
		startId, err := primitive.ObjectIDFromHex(startID)
		if err != nil {
			return nil, "", err
		}

		filter["_id"] = bson.M{"$lt": startId}
	}

	cur, err := collection().Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, "", err
	}

	defer cur.Close(context.TODO())

	if err := cur.All(context.TODO(), &results); err != nil {
		return nil, "", err
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
	filter := bson.M{
		"orgid": orgID,
		"state": bson.M{"$gte": DataRequestStatusProcessedWithoutAction},
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})
	findOptions.SetLimit(int64(limit))

	if startID != "" {
		startId, err := primitive.ObjectIDFromHex(startID)
		if err != nil {
			return nil, "", err
		}

		filter["_id"] = bson.M{"$lt": startId}
	}

	cur, err := collection().Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, "", err
	}

	defer cur.Close(context.TODO())

	if err := cur.All(context.TODO(), &results); err != nil {
		return nil, "", err
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
	filter := bson.M{
		"orgid":  orgID,
		"userid": userID,
		"state":  bson.M{"$lt": DataRequestStatusProcessedWithoutAction},
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})
	findOptions.SetLimit(int64(limit))

	if startID != "" {
		startId, err := primitive.ObjectIDFromHex(startID)
		if err != nil {
			return nil, "", err
		}

		filter["_id"] = bson.M{"$lt": startId}
	}

	cur, err := collection().Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, "", err
	}

	defer cur.Close(context.TODO())

	if err := cur.All(context.TODO(), &results); err != nil {
		return nil, "", err
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
	filter := bson.M{
		"orgid":  orgID,
		"userid": userID,
		"state":  bson.M{"$gte": DataRequestStatusProcessedWithoutAction},
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})
	findOptions.SetLimit(int64(limit))

	if startID != "" {
		startId, err := primitive.ObjectIDFromHex(startID)
		if err != nil {
			return nil, "", err
		}

		filter["_id"] = bson.M{"$lt": startId}
	}

	cur, err := collection().Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, "", err
	}

	defer cur.Close(context.TODO())

	if err := cur.All(context.TODO(), &results); err != nil {
		return nil, "", err
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
	filter := bson.M{
		"orgid":  orgID,
		"userid": userID,
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})
	findOptions.SetLimit(int64(limit))

	if startID != "" {
		startId, err := primitive.ObjectIDFromHex(startID)
		if err != nil {
			return nil, "", err
		}

		filter["_id"] = bson.M{"$lt": startId}
	}

	cur, err := collection().Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, "", err
	}

	defer cur.Close(context.TODO())

	if err := cur.All(context.TODO(), &results); err != nil {
		return nil, "", err
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

	cur, err := collection().Find(context.TODO(), bson.M{"orgid": orgID, "userid": userID, "type": drType})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.TODO())

	if err := cur.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}
