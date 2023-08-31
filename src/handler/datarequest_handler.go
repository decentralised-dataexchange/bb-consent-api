package handler

import (
	dr "github.com/bb-consent/api/src/datarequests"
)

func transformDataReqToResp(dReq dr.DataRequest) dataReqResp {
	return dataReqResp{ID: dReq.ID, UserID: dReq.UserID, UserName: dReq.UserName, OrgID: dReq.OrgID, Type: dReq.Type,
		State: dReq.State, StateStr: dr.GetStatusTypeStr(dReq.State), Comment: dReq.Comments[dReq.State], TypeStr: dr.GetRequestTypeStr(dReq.Type),
		ClosedDate: dReq.ClosedDate.String(), RequestedDate: dReq.ID.Time().String()}
}
