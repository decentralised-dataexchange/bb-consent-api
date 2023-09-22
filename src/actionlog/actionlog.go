package actionlog

import (
	"context"

	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Log type const
const (
	LogTypeSecurity    = 1
	LogTypeAPICalls    = 2
	LogTypeOrgUpdates  = 3
	LogTypeUserUpdates = 4
	LogTypeWebhook     = 5
)

// LogType Log type
type LogType struct {
	ID  int
	Str string
}

// Note: Dont change the ID(s) if new type is needed then add at the end

// LogTypes Array of id and string
var LogTypes = []LogType{
	{ID: LogTypeSecurity, Str: "Security"},
	{ID: LogTypeAPICalls, Str: "API calls"},
	{ID: LogTypeOrgUpdates, Str: "OrgUpdates"},
	{ID: LogTypeUserUpdates, Str: "UserUpdates"},
	{ID: LogTypeWebhook, Str: "Webhooks"}}

// GetTypeStr Get type string from ID
func GetTypeStr(logType int) string {
	return LogTypes[logType-1].Str
}

// ActionLog All access logs
type ActionLog struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Type     int
	TypeStr  string
	OrgID    string
	UserID   string
	UserName string
	Action   string //Free string storing the real log
}

func collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("actionLogs")
}

// Add Adds access log
func Add(log ActionLog) error {
	_, err := collection().InsertOne(context.TODO(), log)
	if err != nil {
		return err
	}
	return nil
}

// GetAccessLogByOrgID gets all notifications of a given user
func GetAccessLogByOrgID(orgID string, startID string, limit int) (results []ActionLog, lastID string, err error) {

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})
	findOptions.SetLimit(int64(limit))

	filter := bson.M{"orgid": orgID}
	if startID != "" {
		startId, err := primitive.ObjectIDFromHex(startID)
		if err != nil {
			return nil, "", err
		}

		filter["_id"] = bson.M{"$lt": startId}
	}

	cursor, err := collection().Find(context.TODO(), filter, findOptions)
	if err != nil {
		return nil, "", err
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
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
