package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	dr "github.com/bb-consent/api/src/datarequests"
	"github.com/bb-consent/api/src/token"
	"github.com/gorilla/mux"
)

func transformDataReqToResp(dReq dr.DataRequest) dataReqResp {
	return dataReqResp{ID: dReq.ID, UserID: dReq.UserID, UserName: dReq.UserName, OrgID: dReq.OrgID, Type: dReq.Type,
		State: dReq.State, StateStr: dr.GetStatusTypeStr(dReq.State), Comment: dReq.Comments[dReq.State], TypeStr: dr.GetRequestTypeStr(dReq.Type),
		ClosedDate: dReq.ClosedDate.String(), RequestedDate: dReq.ID.Time().String()}
}

// GetMyOrgDataRequestStatus Get data request status
func GetMyOrgDataRequestStatus(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["organizationID"]
	userID := token.GetUserID(r)

	startID, limit := common.ParsePaginationQueryParameters(r)
	if limit == 0 {
		limit = 50
	}

	var requestStatus = ""
	requestStatuses, ok := r.URL.Query()["status"]

	if ok {
		requestStatus = requestStatuses[0]
	}

	var err error
	var dReqs []dr.DataRequest
	var lastID string

	if requestStatus == "open" {
		dReqs, lastID, err = dr.GetOpenDataRequestsByOrgUserID(orgID, userID, startID, limit)
	} else if requestStatus == "closed" {
		dReqs, lastID, err = dr.GetClosedDataRequestsByOrgUserID(orgID, userID, startID, limit)
	} else {
		dReqs, lastID, err = dr.GetDataRequestsByOrgUserID(orgID, userID, startID, limit)
	}

	if err != nil {
		m := fmt.Sprintf("Failed to get user: %v data request status for organization: %v", userID, orgID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var drs dataReqResps

	for _, d := range dReqs {
		// checking if atleast one request is ongoing
		if d.State == 1 || d.State == 2 {

			if !drs.IsRequestsOngoing {
				drs.IsRequestsOngoing = true
			}

			if !drs.IsDataDeleteRequestOngoing && d.Type == 1 {
				drs.IsDataDeleteRequestOngoing = true
			}

			if !drs.IsDataDownloadRequestOngoing && d.Type == 2 {
				drs.IsDataDownloadRequestOngoing = true
			}
		}

		drs.DataRequests = append(drs.DataRequests, transformDataReqToResp(d))
	}

	drs.Links = common.CreatePaginationLinks(r, startID, lastID, limit)
	response, _ := json.Marshal(drs)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
