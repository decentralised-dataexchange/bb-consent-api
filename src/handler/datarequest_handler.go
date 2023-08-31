package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	dr "github.com/bb-consent/api/src/datarequests"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/webhooks"
	"github.com/gorilla/mux"
)

// GetDeleteMyData Get my data requests from the organization
func GetDeleteMyData(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgID"]
	userID := token.GetUserID(r)

	drs, err := getDataReqWithUserOrgTypeID(userID, orgID, dr.DataRequestTypeDelete)

	if err != nil {
		m := fmt.Sprintf("Failed to get user: %v data request for organization: %v", userID, orgID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(drs)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

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

func getDataReqWithUserOrgTypeID(userID string, orgID string, typeID int) ([]dataReqResp, error) {
	var err error
	var dReqs []dr.DataRequest
	var drs []dataReqResp

	dReqs, err = dr.GetDataRequestsByUserOrgTypeID(orgID, userID, typeID)
	if err != nil {
		return drs, err
	}

	for _, d := range dReqs {
		drs = append(drs, transformDataReqToResp(d))
	}

	return drs, err
}

type myDataRequestStatus struct {
	RequestOngoing bool
	ID             string
	State          int
	StateStr       string
	RequestedDate  string
}

func getOngoingDataRequest(userID string, orgID string, drType int) (resp myDataRequestStatus, err error) {
	drs, err := getDataReqWithUserOrgTypeID(userID, orgID, drType)

	if err != nil {
		return resp, err
	}

	resp.RequestOngoing = false
	for _, d := range drs {
		if d.State < dr.DataRequestStatusProcessedWithoutAction {
			resp.RequestOngoing = true
			resp.ID = d.ID.Hex()
			resp.State = d.State
			resp.StateStr = d.StateStr
			resp.RequestedDate = d.ID.Time().String()
		}
	}

	return resp, err
}

// DeleteMyData Delete my data from the organization
func DeleteMyData(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgID"]
	userID := token.GetUserID(r)

	resp, err := getOngoingDataRequest(userID, orgID, dr.DataRequestTypeDelete)

	if err == nil && resp.RequestOngoing == true {
		m := fmt.Sprintf("Request (%v) ongoing for user: %v organization: %v", dr.GetRequestTypeStr(dr.DataRequestTypeDelete), userID, orgID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	var dRequest dr.DataRequest
	dRequest.OrgID = orgID
	dRequest.UserID = userID
	dRequest.UserName = token.GetUserName(r)
	dRequest.Type = dr.DataRequestTypeDelete
	dRequest.State = dr.DataRequestStatusInitiated

	dRequest, err = dr.Add(dRequest)
	if err != nil {
		m := fmt.Sprintf("Failed to add data request: %v logs for user: %v organization: %v", dr.GetRequestTypeStr(dr.DataRequestTypeDelete), token.GetUserName(r), orgID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// Trigger webhooks
	go webhooks.TriggerDataRequestWebhookEvent(userID, orgID, dRequest.ID.Hex(), webhooks.EventTypes[webhooks.EventTypeDataDeleteInitiated])

	w.WriteHeader(http.StatusOK)
}