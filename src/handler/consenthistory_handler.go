package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/consenthistory"
	"github.com/bb-consent/api/src/token"
)

type consentHistory struct {
	UserID                 string
	ConsentID              string
	OrgID                  string
	OrgName                string
	PurposeID              string
	PurposeAllowAll        bool
	PurposeName            string
	AttributeID            string
	AttributeDescription   string
	AttributeConsentStatus string
}

type consentHistoryShort struct {
	ID        string
	OrgID     string
	PurposeID string
	Log       string
	TimeStamp string
}

type consentHistoryResp struct {
	ConsentHistory []consentHistoryShort
	Links          common.PaginationLinks
}

func parseConsentHistoryQueryParams(r *http.Request) (startID string, limit int, orgID string, purposeID string, startDate string, endDate string) {
	startID = ""
	orgID = ""
	purposeID = ""
	startDate = ""
	endDate = ""

	startID, limit = common.ParsePaginationQueryParameters(r)

	orgIDs, ok := r.URL.Query()["orgid"]

	if ok {
		orgID = orgIDs[0]
	}

	purposeIDs, ok := r.URL.Query()["purposeid"]

	if ok {
		purposeID = purposeIDs[0]
	}

	startDates, ok := r.URL.Query()["startDate"]

	if ok {
		startDate = startDates[0]
	}

	endDates, ok := r.URL.Query()["endDate"]

	if ok {
		endDate = endDates[0]
	}

	return
}

// GetUserConsentHistory Get user consent history
func GetUserConsentHistory(w http.ResponseWriter, r *http.Request) {
	userID := token.GetUserID(r)

	startID, limit, orgID, purposeID, startDate, endDate := parseConsentHistoryQueryParams(r)
	if limit == 0 {
		limit = 8
	}

	if purposeID != "" && orgID == "" {
		m := fmt.Sprintf("Incorrect filters used. userid: %v orgid: %v should be valid when purposeid: %v is used in query", userID, orgID, purposeID)
		common.HandleError(w, http.StatusBadRequest, m, nil)
		return
	}
	var chs []consenthistory.ConsentHistory
	var lastID = ""
	var err error

	log.Printf("start: %v orgId: %v purposeid: %v limit: %v start:%v end:%v", startID, orgID, purposeID, limit, startDate, endDate)
	if orgID != "" && purposeID != "" {
		sanitizedOrgId := common.Sanitize(orgID)
		sanitizedPurposeId := common.Sanitize(purposeID)

		chs, lastID, err = consenthistory.GetByUserOrgPurposeID(userID, sanitizedOrgId, sanitizedPurposeId, startID, limit)
		if err != nil {
			m := fmt.Sprintf("Failed to get consent history for user id:%v orgID: %v purposeID : %v", userID, orgID, purposeID)
			common.HandleError(w, http.StatusNotFound, m, err)
			return
		}
	} else if orgID != "" {
		sanitizedOrgId := common.Sanitize(orgID)

		chs, lastID, err = consenthistory.GetByUserOrgID(userID, sanitizedOrgId, startID, limit)
		if err != nil {
			m := fmt.Sprintf("Failed to get consent history for user id:%v orgID: %v", userID, orgID)
			common.HandleError(w, http.StatusNotFound, m, err)
			return
		}
	} else if startDate != "" && endDate != "" {
		chs, lastID, err = consenthistory.GetByDateRange(userID, startDate, endDate, startID, limit)
		if err != nil {
			m := fmt.Sprintf("Failed to get consent history for user id:%v start: %v, end: %v", userID, startDate, endDate)
			common.HandleError(w, http.StatusNotFound, m, err)
			return
		}
	} else if orgID == "" && purposeID == "" {
		chs, lastID, err = consenthistory.GetByUserID(userID, startID, limit)
		if err != nil {
			m := fmt.Sprintf("Failed to get consent history for user id:%v", userID)
			common.HandleError(w, http.StatusNotFound, m, err)
			return
		}
	}

	var chsResp consentHistoryResp
	for _, ch := range chs {
		chsResp.ConsentHistory = append(chsResp.ConsentHistory, consentHistoryShort{ID: ch.ID.Hex(), OrgID: ch.OrgID, PurposeID: ch.PurposeID, Log: ch.Log, TimeStamp: ch.ID.Timestamp().Format(time.RFC3339)})
	}

	//chsResp.Links = common.CreatePaginationLinks(r, startID, lastID, limit)
	chsResp.Links = formPaginationLinks(r, startID, lastID, limit, orgID, purposeID)

	response, _ := json.Marshal(chsResp)

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}

func formPaginationLinks(r *http.Request, startID string, lastID string, limit int, orgID string, purposeID string) (pagination common.PaginationLinks) {
	l := common.CreatePaginationLinks(r, startID, lastID, limit)

	pagination = l
	if orgID != "" && purposeID != "" {
		pagination.Self = fmt.Sprintf("%s&orgid=%s&purposeid=%s", l.Self, orgID, purposeID)
		if l.Next != "" {
			pagination.Next = fmt.Sprintf("%s&orgid=%s&purposeid=%s", l.Next, orgID, purposeID)
		}
	} else if orgID != "" {
		pagination.Self = fmt.Sprintf("%s&orgid=%s", l.Self, orgID)
		if l.Next != "" {
			pagination.Next = fmt.Sprintf("%s&orgid=%s", l.Next, orgID)
		}
	}
	return
}

func consentHistoryPurposeAdd(ch consentHistory) error {
	var c consenthistory.ConsentHistory

	c.ConsentID = ch.ConsentID
	c.UserID = ch.UserID
	c.OrgID = ch.OrgID
	c.PurposeID = ch.PurposeID

	var val = "DisAllow"
	if ch.PurposeAllowAll == true {
		val = "Allow"
	}
	c.Log = fmt.Sprintf("Updated consent value to <%s> for the purpose <%s> in organization <%s>",
		val, ch.PurposeName, ch.OrgName)

	log.Printf("The log is: %s", c.Log)
	_, err := consenthistory.Add(c)
	if err != nil {
		return err
	}

	return nil
}

func consentHistoryAttributeAdd(ch consentHistory) error {
	var c consenthistory.ConsentHistory

	c.ConsentID = ch.ConsentID
	c.UserID = ch.UserID
	c.OrgID = ch.OrgID
	c.PurposeID = ch.PurposeID

	c.Log = fmt.Sprintf("Updated consent value to <%s> for attribute <%s> in organization <%s> for the purpose <%s>",
		ch.AttributeConsentStatus, ch.AttributeDescription, ch.OrgName, ch.PurposeName)

	log.Printf("The log is: %s", c.Log)
	_, err := consenthistory.Add(c)
	if err != nil {
		return err
	}

	return nil
}
