package actionlog

import (
	"fmt"
	"log"
)

func doActionLog(l ActionLog) error {
	err := Add(l)
	if err != nil {
		log.Printf("Failed to add access log for user: %v org: %v", l.UserID, l.OrgID)
	}
	fmt.Printf("%v \n", l)
	return err
}

// LogOrgSecurityCalls Logs all important access related entries
func LogOrgSecurityCalls(userID string, uName string, orgID string, aLog string) {
	var l ActionLog
	l.OrgID = orgID
	l.UserID = userID
	l.UserName = uName
	l.Action = aLog
	l.Type = LogTypeSecurity
	l.TypeStr = GetTypeStr(l.Type)

	doActionLog(l)
}

// LogOrgWebhookCalls Logs all webhook triggers based on different events
func LogOrgWebhookCalls(userID string, uName string, orgID string, aLog string) {
	var l ActionLog
	l.OrgID = orgID
	l.UserID = userID
	l.UserName = uName
	l.Action = aLog
	l.Type = LogTypeWebhook
	l.TypeStr = GetTypeStr(l.Type)

	doActionLog(l)
}
