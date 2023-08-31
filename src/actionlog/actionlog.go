package actionlog

import (
	"github.com/bb-consent/api/src/database"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
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
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Type     int
	TypeStr  string
	OrgID    string
	UserID   string
	UserName string
	Action   string //Free string storing the real log
}

func session() *mgo.Session {
	return database.DB.Session.Copy()
}

func collection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("actionLogs")
}

// Add Adds access log
func Add(log ActionLog) error {
	s := session()
	defer s.Close()

	err := collection(s).Insert(log)
	if err != nil {
		return err
	}
	return nil
}

// GetAccessLogByOrgID gets all notifications of a given user
func GetAccessLogByOrgID(orgID string, startID string, limit int) (results []ActionLog, lastID string, err error) {
	s := session()
	defer s.Close()

	if startID == "" {
		err = collection(s).Find(bson.M{"orgid": orgID}).Sort("-_id").Limit(limit).All(&results)
	} else {
		err = collection(s).Find(bson.M{"orgid": orgID, "_id": bson.M{"$lt": bson.ObjectIdHex(startID)}}).Sort("-_id").Limit(limit).All(&results)
	}

	lastID = ""
	if err == nil {
		if len(results) != 0 && len(results) == (limit) {
			lastID = results[len(results)-1].ID.Hex()
		}
	}

	return results, lastID, err
}
