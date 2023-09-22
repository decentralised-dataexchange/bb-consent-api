package consenthistory

import (
	"context"
	"time"

	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConsentHistory HOlds the consent logs
type ConsentHistory struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    string
	OrgID     string
	PurposeID string
	ConsentID string
	Log       string
}

func collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("consentHistory")
}

// Add Adds a consent history to the collection
func Add(ch ConsentHistory) (ConsentHistory, error) {

	ch.ID = primitive.NewObjectID()

	_, err := collection().InsertOne(context.TODO(), &ch)

	return ch, err
}

// GetByUserID Gets all history of a given userID
func GetByUserID(userID string, startID string, limit int) ([]ConsentHistory, string, error) {
	filter := bson.M{
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

	var results []ConsentHistory

	cur, err := collection().Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, "", err
	}

	defer cur.Close(context.TODO())

	if err := cur.All(context.TODO(), &results); err != nil {
		return nil, "", err
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

	filter := bson.M{"userid": userID, "orgid": orgID, "purposeid": purposeID}
	options := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})

	var result ConsentHistory
	err := collection().FindOne(context.TODO(), filter, options).Decode(&result)

	return result, err
}

// GetByUserOrgPurposeID Gets all history of a given userID in an organization with purposeID
func GetByUserOrgPurposeID(userID string, orgID string, purposeID string, startID string, limit int) ([]ConsentHistory, string, error) {

	filter := bson.M{
		"userid":    userID,
		"orgid":     orgID,
		"purposeid": purposeID,
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

	var results []ConsentHistory

	cur, err := collection().Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, "", err
	}

	defer cur.Close(context.TODO())

	if err := cur.All(context.TODO(), &results); err != nil {
		return nil, "", err
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
	filter := bson.M{
		"userid": userID,
		"orgid":  orgID,
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

	var results []ConsentHistory

	cur, err := collection().Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, "", err
	}

	defer cur.Close(context.TODO())

	if err := cur.All(context.TODO(), &results); err != nil {
		return nil, "", err
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

	sID := primitive.NewObjectIDFromTimestamp(sDate)
	eID := primitive.NewObjectIDFromTimestamp(eDate)

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})
	findOptions.SetLimit(int64(limit))

	var cur *mongo.Cursor

	if startID == "" {
		cur, err = collection().Find(context.TODO(), bson.M{"userid": userID, "_id": bson.M{"$gte": sID, "$lt": eID}}, findOptions)
		if err != nil {
			return nil, "", err
		}
	} else {
		startId, err := primitive.ObjectIDFromHex(startID)
		if err != nil {
			return nil, "", err
		}

		cur, err = collection().Find(context.TODO(), bson.M{"userid": userID, "_id": bson.M{"$lt": startId, "$gte": sID}}, findOptions)
		if err != nil {
			return nil, "", err
		}
	}

	defer cur.Close(context.TODO())

	if err := cur.All(context.TODO(), &results); err != nil {
		return nil, "", err
	}

	var lastID = ""
	if err == nil {
		if len(results) != 0 && len(results) == (limit) {
			lastID = results[len(results)-1].ID.Hex()
		}
	}

	return results, lastID, err
}
