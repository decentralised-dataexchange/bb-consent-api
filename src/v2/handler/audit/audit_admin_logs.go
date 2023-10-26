package audit

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/actionlog"
	"github.com/bb-consent/api/src/v2/paginate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ListActionLogsError is an error enumeration for action logs API.
type ListActionLogsError int

const (
	// ActionLogTypeIsMissingError indicates that the action log query param is missing.
	ActionLogTypeIsMissingError ListActionLogsError = iota
)

// Error returns the string representation of the error.
func (e ListActionLogsError) Error() string {
	switch e {
	case ActionLogTypeIsMissingError:
		return "Query param action log type is missing!"
	default:
		return "Unknown error!"
	}
}

// ParseListActionLogQueryParams parses query params for listing action logs.
func ParseListActionLogQueryParams(r *http.Request) (int, error) {
	query := r.URL.Query()
	var logType int

	// Check if logType query param is provided.
	if r, ok := query["logType"]; ok && len(r) > 0 {
		if oInt, err := strconv.Atoi(r[0]); err == nil && oInt >= 1 {
			logType = oInt
			return logType, nil
		}
	}

	return logType, ActionLogTypeIsMissingError
}

type listActionLogsResp struct {
	ActionLogs interface{}         `json:"logs"`
	Pagination paginate.Pagination `json:"pagination"`
}

func returnHTTPResponse(resp interface{}, w http.ResponseWriter) {
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// AuditGetOrgLogs Get action logs for the organization
func AuditGetOrgLogs(w http.ResponseWriter, r *http.Request) {
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

	// Repository
	actionLogRepo := actionlog.ActionLogRepository{}
	actionLogRepo.Init(organisationId)

	var pipeline []primitive.M

	logType, err := ParseListActionLogQueryParams(r)
	if err != nil && errors.Is(err, ActionLogTypeIsMissingError) {
		pipeline = []bson.M{{"$sort": bson.M{"timestamp": -1}}}
	} else {
		pipeline = []bson.M{{"$match": bson.M{"type": logType}}, {"$sort": bson.M{"timestamp": -1}}}
	}
	// Return all action logs
	var actionLogs []actionlog.ActionLog
	query := paginate.PaginateDBObjectsQueryUsingPipeline{
		Pipeline:   pipeline,
		Collection: actionlog.Collection(),
		Context:    context.Background(),
		Limit:      limit,
		Offset:     offset,
	}
	result, err := paginate.PaginateDBObjectsUsingPipeline(query, &actionLogs)
	if err != nil {
		if errors.Is(err, paginate.EmptyDBError) {
			emptyActionLogs := make([]interface{}, 0)
			resp := listActionLogsResp{
				ActionLogs: emptyActionLogs,
				Pagination: result.Pagination,
			}
			returnHTTPResponse(resp, w)
			return
		}
		m := "Failed to paginate action log"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return

	}
	resp := listActionLogsResp{
		ActionLogs: result.Items,
		Pagination: result.Pagination,
	}
	returnHTTPResponse(resp, w)

}
