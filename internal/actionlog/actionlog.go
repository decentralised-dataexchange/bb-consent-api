package actionlog

import (
	"context"
	"time"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/database"
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
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Type      int                `json:"type"`
	TypeStr   string             `json:"typeStr"`
	OrgID     string             `json:"orgId"`
	UserID    string             `json:"userId"`
	UserName  string             `json:"userName"`
	Action    string             `json:"action"` //Free string storing the real log
	Timestamp string             `json:"timestamp"`
}

type ActionLogRepository struct {
	DefaultFilter bson.M
}

// Init
func (actionLogRepo *ActionLogRepository) Init(organisationId string) {
	actionLogRepo.DefaultFilter = bson.M{"orgid": organisationId}
}

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("actionLogs")
}

// Add Adds access log
func Add(log ActionLog) error {
	log.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	_, err := Collection().InsertOne(context.TODO(), log)
	if err != nil {
		return err
	}
	return nil
}

// GetAccessLogByOrgID gets all notifications of a given user
func (actionLogRepo *ActionLogRepository) GetAccessLogByOrgID() (results []ActionLog, err error) {
	filter := actionLogRepo.DefaultFilter

	cursor, err := Collection().Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, err
}

// Count logs
func (actionLogRepo *ActionLogRepository) CountLogs() (int64, error) {
	filter := actionLogRepo.DefaultFilter

	count, err := Collection().CountDocuments(context.Background(), filter)
	if err != nil {
		return count, nil
	}

	return count, nil
}

// GetLogOfIndexHundread
func (actionLogRepo *ActionLogRepository) GetLogOfIndexHundread() (ActionLog, error) {

	var result ActionLog
	filter := actionLogRepo.DefaultFilter
	opts := options.FindOne().SetSort(bson.M{"timestamp": -1}).SetSkip(100)
	err := Collection().FindOne(context.TODO(), filter, opts).Decode(&result)
	if err != nil {
		return result, err
	}

	return result, err
}

// DeleteLogsLessThanTimestamp
func (actionLogRepo *ActionLogRepository) DeleteLogsLessThanTimestamp(timestamp string) error {

	filter := common.CombineFilters(actionLogRepo.DefaultFilter, bson.M{
		"timestamp": bson.M{"$lt": timestamp},
	})

	_, err := Collection().DeleteMany(context.Background(), filter)
	if err != nil {
		return err
	}

	return nil
}
