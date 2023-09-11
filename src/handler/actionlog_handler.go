package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/actionlog"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/gorilla/mux"
)

type orgLog struct {
	ID        string
	Type      int
	TypeStr   string
	UserID    string
	UserName  string
	TimeStamp string
	Log       string
}
type orgLogsResp struct {
	Logs  []orgLog
	Links common.PaginationLinks
}

// GetOrgLogs Get action logs for the organization
func GetOrgLogs(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgID"]

	startID, limit := common.ParsePaginationQueryParameters(r)
	if limit == 0 {
		limit = 50
	}

	logs, lastID, err := actionlog.GetAccessLogByOrgID(orgID, startID, limit)
	if err != nil {
		m := fmt.Sprintf("Failed to get logs for organization: %v", orgID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var ls orgLogsResp
	for _, l := range logs {
		ls.Logs = append(ls.Logs, orgLog{ID: l.ID.Hex(), Type: l.Type, TypeStr: l.TypeStr,
			UserID: l.UserID, UserName: l.UserName, TimeStamp: l.ID.Time().String(), Log: l.Action})
	}

	ls.Links = common.CreatePaginationLinks(r, startID, lastID, limit)
	response, _ := json.Marshal(ls)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}
