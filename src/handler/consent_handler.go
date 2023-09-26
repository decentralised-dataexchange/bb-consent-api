package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/actionlog"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/consent"
	"github.com/bb-consent/api/src/consenthistory"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/user"
	"github.com/bb-consent/api/src/webhooks"
	"github.com/gorilla/mux"
)

type consentStatus struct {
	Consented string
	TimeStamp time.Time
	Days      int
	Remaining int
}

// ConsentCount Counts the total consent attributes and consented ones
type ConsentCount struct {
	Total     int
	Consented int
}
type consentResp struct {
	ID          string
	Description string
	Value       string
	Status      consentStatus
}
type consentsAndPurpose struct {
	Purpose  org.Purpose
	Count    ConsentCount
	Consents []consentResp
}

// ConsentsResp Consent response struct definition
type ConsentsResp struct {
	ID                  string
	OrgID               string
	UserID              string
	ConsentsAndPurposes []consentsAndPurpose
}

// DataRetentionPolicyResp Data retention policy response struct definition
type DataRetentionPolicyResp struct {
	Expiry string
}

type consentsAndPurposeWithDataRetention struct {
	Purpose       org.Purpose
	Count         ConsentCount
	Consents      []consentResp
	DataRetention DataRetentionPolicyResp
}

// ConsentsRespWithDataRetention Consent response struct definition with data retention for each purpose
type ConsentsRespWithDataRetention struct {
	ID                  string
	OrgID               string
	UserID              string
	ConsentsAndPurposes []consentsAndPurposeWithDataRetention
}

// GetConsentResponse Gets all consents and formulates the response
func GetConsentResponse(w http.ResponseWriter, userID string, orgID string) (ConsentsResp, error) {
	var c ConsentsResp
	o, err := org.Get(orgID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization user: %v org: %v", userID, orgID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return c, err
	}

	sanitizedOrgId := common.Sanitize(orgID)
	sanitizedUserId := common.Sanitize(userID)

	consents, err := consent.GetByUserOrg(sanitizedUserId, sanitizedOrgId)
	if err != nil {
		if err.Error() == "not found" {
			var con consent.Consents
			con.OrgID = orgID
			con.UserID = userID
			consents, err = consent.Add(con)
			if err != nil {
				m := fmt.Sprintf("Failed to fetch consents user: %v org: %v", userID, orgID)
				common.HandleError(w, http.StatusInternalServerError, m, err)
				return c, err
			}
		} else {
			m := fmt.Sprintf("Failed to fetch consents user: %v org: %v", userID, orgID)
			common.HandleError(w, http.StatusInternalServerError, m, err)
			return c, err
		}
	}
	c = createConsentGetResponse(consents, o)
	c.ID = consents.ID.Hex()
	c.OrgID = orgID
	c.UserID = userID

	return c, nil
}

// GetConsents Gets all consent entries in the collection
func GetConsents(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgID"]
	userID := mux.Vars(r)["userID"]

	// Fetching the organisation by ID
	o, err := org.Get(orgID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization for user: %v org: %v", userID, orgID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	sanitizedOrgId := common.Sanitize(orgID)
	sanitizedUserId := common.Sanitize(userID)

	c, err := GetConsentResponse(w, sanitizedUserId, sanitizedOrgId)
	if err != nil {
		log.Printf("Failed to get consents for user: %v org: %v err: %v", userID, orgID, err)
		return
	}

	// For holding the API response
	var RespData ConsentsRespWithDataRetention

	// Constructing the response data
	RespData.ID = c.ID
	RespData.UserID = c.UserID
	RespData.OrgID = c.OrgID

	// Constructing the response data
	if len(c.ConsentsAndPurposes) > 0 {
		for _, tempConsentsAndPurpose := range c.ConsentsAndPurposes {

			var tempConsentsAndPurposes consentsAndPurposeWithDataRetention
			tempConsentsAndPurposes.Purpose = tempConsentsAndPurpose.Purpose
			tempConsentsAndPurposes.Count = tempConsentsAndPurpose.Count

			if tempConsentsAndPurpose.Count.Total > 0 {
				if tempConsentsAndPurpose.Consents == nil {
					tempConsentsAndPurposes.Consents = make([]consentResp, 0)
				} else {
					tempConsentsAndPurposes.Consents = tempConsentsAndPurpose.Consents
				}

				RespData.ConsentsAndPurposes = append(RespData.ConsentsAndPurposes, tempConsentsAndPurposes)
			}

		}

	}

	// Check if data retention policy enabled for the organisation
	if o.DataRetention.Enabled {

		if len(RespData.ConsentsAndPurposes) > 0 {
			// Add data retention expiry for each purpose if available
			for i, _ := range RespData.ConsentsAndPurposes {

				// Check if purpose has consent as lawful basis
				if !RespData.ConsentsAndPurposes[i].Purpose.LawfulUsage {

					// Check if the purpose is allowed
					if RespData.ConsentsAndPurposes[i].Count.Consented > 0 {

						latestConsentHistory, err := consenthistory.GetLatestByUserOrgPurposeID(RespData.UserID, RespData.OrgID, RespData.ConsentsAndPurposes[i].Purpose.ID)
						if err != nil {
							continue
						}

						RespData.ConsentsAndPurposes[i].DataRetention.Expiry = latestConsentHistory.ID.Timestamp().Add(time.Second * time.Duration(o.DataRetention.RetentionPeriod)).UTC().String()
						log.Printf("Expiry for purpose:%v is %v", RespData.ConsentsAndPurposes[i].Purpose.ID, RespData.ConsentsAndPurposes[i].DataRetention.Expiry)
					}

				}

			}
		}

	}

	//fmt.Printf("c:%v", c)
	response, _ := json.Marshal(RespData)

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}

// GetConsentByID Gets a single consent by given id
func GetConsentByID(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgID"]
	userID := mux.Vars(r)["userID"]
	consentID := mux.Vars(r)["consentID"]

	o, err := org.Get(orgID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization for user: %v org: %v", userID, orgID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	consent, err := consent.Get(consentID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch consensts by ID: %v \n", consentID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	c := createConsentGetResponse(consent, o)
	response, _ := json.Marshal(c)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}

type consentPurposeResp struct {
	ID            string
	ConsentID     string
	OrgID         string
	UserID        string
	Consents      consentsAndPurpose
	DataRetention DataRetentionPolicyResp
}

// GetConsentPurposeByID Gets all the consents for agiven purpose by ID
func GetConsentPurposeByID(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgID"]
	userID := mux.Vars(r)["userID"]
	consentID := mux.Vars(r)["consentID"]
	purposeID := mux.Vars(r)["purposeID"]

	o, err := org.Get(orgID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization for user: %v org: %v", userID, orgID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	consent, err := consent.Get(consentID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch consensts by ID: %v for user: %v", consentID, userID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	c := createConsentGetResponse(consent, o)

	var found = 0
	var cp consentsAndPurpose
	for _, item := range c.ConsentsAndPurposes {
		if item.Purpose.ID == purposeID {
			cp = item
			found++
		}
	}
	if found == 0 {
		//TODO: Handle the case where the purpose ID is non existent
	}

	var cpResp consentPurposeResp
	cpResp.ID = purposeID
	cpResp.ConsentID = consentID
	cpResp.OrgID = orgID
	cpResp.UserID = userID
	cpResp.Consents = cp

	sanitizedOrgId := common.Sanitize(orgID)
	sanitizedUserId := common.Sanitize(userID)
	sanitizedPurposeId := common.Sanitize(purposeID)

	// Data retention expiry
	if o.DataRetention.Enabled {

		// Check if atleast one attribute consent is allowed
		isPurposeAllowed := false
		for _, attributeConsent := range cpResp.Consents.Consents {
			if attributeConsent.Status.Consented == common.ConsentStatusAllow {
				isPurposeAllowed = true
				break
			}
		}

		if isPurposeAllowed {

			latestConsentHistory, err := consenthistory.GetLatestByUserOrgPurposeID(sanitizedUserId, sanitizedOrgId, sanitizedPurposeId)
			if err != nil {
				response, _ := json.Marshal(cpResp)
				w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
				w.Write(response)
				return
			}

			cpResp.DataRetention.Expiry = latestConsentHistory.ID.Timestamp().Add(time.Second * time.Duration(o.DataRetention.RetentionPeriod)).UTC().String()
		}
	}

	response, _ := json.Marshal(cpResp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}

// GetAllUsersConsentedToAttribute Gets all users who conseted to a given attribute
func GetAllUsersConsentedToAttribute(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgID"]
	purposeID := mux.Vars(r)["purposeID"]
	attributeID := mux.Vars(r)["attributeID"]

	aLog := fmt.Sprintf("Organization API: %v called by user: %v", r.URL.Path, token.GetUserName(r))
	actionlog.LogOrgAPICalls(token.GetUserID(r), token.GetUserName(r), orgID, aLog)

	startID, limit := common.ParsePaginationQueryParameters(r)
	if limit == 0 {
		limit = 50
	}

	purpose, err := org.GetPurpose(orgID, purposeID)
	if err != nil {
		m := fmt.Sprintf("Failed to locate purposeID: %v for orgID: %v", orgID, purposeID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	// If the purpose is lawful usage then we can fetch all the subscribed users right away.
	if purpose.LawfulUsage == true {
		users, lastID, err := user.GetOrgSubscribeUsers(orgID, startID, limit)
		if err != nil {
			m := fmt.Sprintf("Failed to get user subscribed to organization :%v", orgID)
			common.HandleError(w, http.StatusNotFound, m, err)
			return
		}

		var ou orgUsers
		for _, u := range users {
			ou.Users = append(ou.Users, orgUser{ID: u.ID.Hex(), Name: u.Name, Phone: u.Phone, Email: u.Email})
		}

		ou.Links = common.CreatePaginationLinks(r, startID, lastID, limit)
		response, _ := json.Marshal(ou)
		w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
		w.Write(response)
		return
	}

	sanitizedOrgId := common.Sanitize(orgID)
	sanitizedPurposeId := common.Sanitize(purposeID)
	sanitizedAttributeId := common.Sanitize(attributeID)
	sanitizedStartId := common.Sanitize(startID)

	userIDs, nextID, err := consent.GetConsentedUsers(sanitizedOrgId, sanitizedPurposeId, sanitizedAttributeId, sanitizedStartId, limit)

	if err != nil {
		m := fmt.Sprintf("Failed to fetch users constented orgID: %v purposeID: %v attributeID: %v", orgID, purposeID, attributeID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	var resp orgUsers
	for _, userID := range userIDs {
		u, err := user.Get(userID)
		if err != nil {
			//TODO: This is unexpected! report error here?
			continue
		}
		resp.Users = append(resp.Users, orgUser{ID: u.ID.Hex(), Name: u.Name, Phone: u.Phone, Email: u.Email})
	}

	resp.Links = common.CreatePaginationLinks(r, startID, nextID, limit)
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}

// GetPurposeAllConsentStatus Get all consent attributes status of a given purpose
func GetPurposeAllConsentStatus(w http.ResponseWriter, r *http.Request) {
	consentID := mux.Vars(r)["consentID"]
	userID := mux.Vars(r)["userID"]
	purposeID := mux.Vars(r)["purposeID"]

	c, err := consent.Get(consentID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch consensts by ID: %v for user: %v", consentID, userID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	var consents []consent.Consent
	for _, p := range c.Purposes {
		if p.ID == purposeID {
			consents = p.Consents
		}
	}

	var consentStatus = common.ConsentStatusDisAllow
	for _, cons := range consents {
		if cons.Status.Consented == common.ConsentStatusAllow {
			consentStatus = common.ConsentStatusAllow
			break
		}
		if cons.Status.Days != 0 {
			remaining := cons.Status.Days - int((time.Now().Sub(cons.Status.TimeStamp).Hours())/24)
			if remaining <= 0 {
				consentStatus = common.ConsentStatusAllow
				break
			}
		}
	}

	type purposeStatus struct {
		Consented string
	}

	response, _ := json.Marshal(purposeStatus{consentStatus})
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}

// GetAllUsersConsentedToPurpose Gets all users who conseted to a given purpose
func GetAllUsersConsentedToPurpose(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgID"]
	purposeID := mux.Vars(r)["purposeID"]

	aLog := fmt.Sprintf("Organization API: %v called by user: %v", r.URL.Path, token.GetUserName(r))
	actionlog.LogOrgAPICalls(token.GetUserID(r), token.GetUserName(r), orgID, aLog)

	startID, limit := common.ParsePaginationQueryParameters(r)
	if limit == 0 {
		limit = 50
	}

	purpose, err := org.GetPurpose(orgID, purposeID)
	if err != nil {
		m := fmt.Sprintf("Failed to locate purposeID: %v for orgID: %v", orgID, purposeID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	// If the purpose is lawful usage then we can fetch all the subscribed users right away.
	//TODO: Move it as a function
	if purpose.LawfulUsage == true {
		users, lastID, err := user.GetOrgSubscribeUsers(orgID, startID, limit)
		if err != nil {
			m := fmt.Sprintf("Failed to get user subscribed to organization :%v", orgID)
			common.HandleError(w, http.StatusNotFound, m, err)
			return
		}

		var ou orgUsers
		for _, u := range users {
			ou.Users = append(ou.Users, orgUser{ID: u.ID.Hex(), Name: u.Name, Phone: u.Phone, Email: u.Email})
		}

		ou.Links = common.CreatePaginationLinks(r, startID, lastID, limit)
		response, _ := json.Marshal(ou)
		w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
		w.Write(response)
		return
	}
	sanitizedOrgId := common.Sanitize(orgID)
	sanitizedPurposeId := common.Sanitize(purposeID)
	sanitizedStartId := common.Sanitize(startID)

	userIDs, nextID, err := consent.GetPurposeConsentedAllUsers(sanitizedOrgId, sanitizedPurposeId, sanitizedStartId, limit)

	if err != nil {
		m := fmt.Sprintf("Failed to fetch users constented orgID: %v purposeID: %v ", orgID, purposeID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	var resp orgUsers
	for _, userID := range userIDs {
		u, err := user.Get(userID)
		if err != nil {
			//TODO: This is unexpected! report error here?
			continue
		}
		resp.Users = append(resp.Users, orgUser{ID: u.ID.Hex(), Name: u.Name, Phone: u.Phone, Email: u.Email})
	}

	resp.Links = common.CreatePaginationLinks(r, startID, nextID, limit)
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}

type purposeAllUpReq struct {
	Consented string `valid:"required"`
}

// UpdatePurposeAllConsentsv2 Updates all consent attributes of a given purpose
func UpdatePurposeAllConsentsv2(w http.ResponseWriter, r *http.Request) {

	var purposeUp purposeAllUpReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &purposeUp)

	// validating request payload
	valid, err := govalidator.ValidateStruct(purposeUp)
	if !valid {
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	consentID := mux.Vars(r)["consentID"]
	orgID := mux.Vars(r)["orgID"]
	userID := mux.Vars(r)["userID"]
	purposeID := mux.Vars(r)["purposeID"]

	// Validating the purpose consent value
	purposeUp.Consented = strings.ToLower(purposeUp.Consented)
	switch purposeUp.Consented {

	case strings.ToLower(common.ConsentStatusAskMe):
		purposeUp.Consented = common.ConsentStatusAskMe
	case strings.ToLower(common.ConsentStatusAllow):
		purposeUp.Consented = common.ConsentStatusAllow
	case strings.ToLower(common.ConsentStatusDisAllow):
		purposeUp.Consented = common.ConsentStatusDisAllow
	default:
		m := fmt.Sprintf("Please provide a valid value for consent; Failed to update purpose consent: %v for org: %v user: %v", consentID, orgID, userID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	o, err := org.Get(orgID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization for user: %v org: %v", userID, orgID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	c, err := consent.Get(consentID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch consensts by ID: %v for user: %v", consentID, userID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	// Combine org and consent details to get unified view
	cResp := createConsentGetResponse(c, o)

	var found = 0
	var cp consentsAndPurpose
	for _, item := range cResp.ConsentsAndPurposes {
		if item.Purpose.ID == purposeID {
			cp = item
			found++
		}
	}
	if found == 0 {
		//TODO: Handle the case where the purpose ID is non existent
	}

	validBasis := cp.Purpose.LawfulBasisOfProcessing == org.ConsentBasis || cp.Purpose.LawfulBasisOfProcessing == org.LegitimateInterestBasis

	if !validBasis {
		errorMsg := fmt.Sprintf("Invalid lawfull basis for purpose: %v, org: %v, user: %v", purposeID, orgID, userID)
		common.HandleError(w, http.StatusBadRequest, errorMsg, err)
		return
	}

	//TODO: HAckish, not optimized at all
	var cnew consent.Consents
	cnew.ID = c.ID
	cnew.OrgID = c.OrgID
	cnew.UserID = c.UserID
	cnew.Purposes = nil

	for _, e := range c.Purposes {
		if e.ID != purposeID {
			cnew.Purposes = append(cnew.Purposes, e)
		}
	}

	var purposeConsentStatus = false
	if purposeUp.Consented == common.ConsentStatusAllow {
		purposeConsentStatus = true
	}

	var purpose consent.Purpose
	purpose.ID = purposeID
	purpose.AllowAll = purposeConsentStatus

	for _, e := range cp.Consents {
		var conNew consent.Consent
		conNew.TemplateID = e.ID
		conNew.Value = e.Value
		conNew.Status.Consented = purposeUp.Consented
		conNew.Status.Days = 0

		purpose.Consents = append(purpose.Consents, conNew)
	}

	cnew.Purposes = append(cnew.Purposes, purpose)
	c, err = consent.UpdatePurposes(cnew)

	if err != nil {
		m := fmt.Sprintf("Failed to update consent:%v for org: %v user: %v", cnew, orgID, userID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	cResp = createConsentGetResponse(c, o)

	var ch consentHistory
	ch.UserID = userID
	ch.OrgID = orgID
	ch.OrgName = o.Name
	ch.PurposeID = purposeID
	ch.PurposeName = getPurposeFromID(o.Purposes, purposeID).Name
	ch.ConsentID = c.ID.Hex()
	ch.PurposeAllowAll = false
	ch.PurposeAllowAll = purposeConsentStatus

	purpose.AllowAll = purposeConsentStatus
	err = consentHistoryPurposeAdd(ch)
	if err != nil {
		m := fmt.Sprintf("Failed to update log for consent: %v for org: %v user: %v ", consentID, orgID, userID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var cRespWithDataRetention ConsentsRespWithDataRetention
	cRespWithDataRetention.OrgID = cResp.OrgID
	cRespWithDataRetention.UserID = cResp.UserID
	cRespWithDataRetention.ID = cResp.ID

	sanitizedOrgId := common.Sanitize(orgID)
	sanitizedUserId := common.Sanitize(userID)

	for i, _ := range cResp.ConsentsAndPurposes {
		var tempConsentsAndPurposeWithDataRetention consentsAndPurposeWithDataRetention
		tempConsentsAndPurposeWithDataRetention.Consents = cResp.ConsentsAndPurposes[i].Consents
		tempConsentsAndPurposeWithDataRetention.Count = cResp.ConsentsAndPurposes[i].Count
		tempConsentsAndPurposeWithDataRetention.Purpose = cResp.ConsentsAndPurposes[i].Purpose

		if o.DataRetention.Enabled {

			// Check if purpose is allowed
			if cResp.ConsentsAndPurposes[i].Count.Consented > 0 {
				latestConsentHistory, err := consenthistory.GetLatestByUserOrgPurposeID(sanitizedUserId, sanitizedOrgId, cResp.ConsentsAndPurposes[i].Purpose.ID)
				if err != nil {
					cRespWithDataRetention.ConsentsAndPurposes = append(cRespWithDataRetention.ConsentsAndPurposes, tempConsentsAndPurposeWithDataRetention)
					continue
				}

				tempConsentsAndPurposeWithDataRetention.DataRetention.Expiry = latestConsentHistory.ID.Timestamp().Add(time.Second * time.Duration(o.DataRetention.RetentionPeriod)).UTC().String()
			}
		}

		cRespWithDataRetention.ConsentsAndPurposes = append(cRespWithDataRetention.ConsentsAndPurposes, tempConsentsAndPurposeWithDataRetention)

	}

	// Trigger webhooks
	var consentedAttributes []string
	for _, pConsent := range purpose.Consents {
		consentedAttributes = append(consentedAttributes, pConsent.TemplateID)
	}

	webhookEventTypeID := webhooks.EventTypeConsentDisAllowed
	if purposeUp.Consented == common.ConsentStatusAllow {
		webhookEventTypeID = webhooks.EventTypeConsentAllowed
	}

	go webhooks.TriggerConsentWebhookEvent(userID, purposeID, consentID, orgID, webhooks.EventTypes[webhookEventTypeID], strconv.FormatInt(time.Now().UTC().Unix(), 10), 0, consentedAttributes)

	response, _ := json.Marshal(cRespWithDataRetention)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}

func getPurposeTemplates(purposeID string, purposes []org.Purpose, templates []org.Template) []string {
	var pt []string
	for _, t := range templates {
		for _, pID := range t.PurposeIDs {
			if pID == purposeID {
				pt = append(pt, t.ID)
			}
		}
	}
	return pt
}

func getPurposeConsentStatus(pt []string, cp consent.Purpose) bool {
	if len(pt) != len(cp.Consents) {
		return false
	}
	for _, c := range cp.Consents {
		if c.Status.Consented == common.ConsentStatusDisAllow {
			return false
		}
		if c.Status.Consented == common.ConsentStatusAskMe {
			if c.Status.Days != 0 {
				remainingDays := c.Status.Days - int((time.Now().Sub(c.Status.TimeStamp).Hours())/24)
				if remainingDays <= 0 {
					return false
				}
			}
		}
	}
	return true
}

func updatePurposeStatus(purposeID string, purposes []org.Purpose, templates []org.Template, consents *consent.Consents) {
	for i := range consents.Purposes {
		if consents.Purposes[i].ID != purposeID {
			continue
		}
		pt := getPurposeTemplates(purposeID, purposes, templates)
		consents.Purposes[i].AllowAll = getPurposeConsentStatus(pt, consents.Purposes[i])
	}
}

type consentUpdateReq struct {
	Consented string `valid:"required"`
	Days      int
}

// UpdatePurposeAttribute Updated one single consent attribute in a purpose
func UpdatePurposeAttribute(w http.ResponseWriter, r *http.Request) {

	var consentUp consentUpdateReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &consentUp)

	// validating request payload
	valid, err := govalidator.ValidateStruct(consentUp)
	if !valid {
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	consentID := mux.Vars(r)["consentID"]
	attributeID := mux.Vars(r)["attributeID"]

	log.Printf("req: %v", consentUp)
	orgID := mux.Vars(r)["orgID"]
	userID := mux.Vars(r)["userID"]

	c, err := consent.Get(consentID)
	if err != nil {
		m := fmt.Sprintf("Failed to get consent by ID: %v", consentID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	if c.OrgID != orgID || c.UserID != userID {
		m := fmt.Sprintf("Consent: %v does not belong to org: %v user: %v ", consentID, orgID, userID)
		common.HandleError(w, http.StatusForbidden, m, err)
		return
	}

	o, err := org.Get(orgID)
	if err != nil {
		m := fmt.Sprintf("Faile to get org: %v", consentID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	purposeID := mux.Vars(r)["purposeID"]

	// Validating the consent value
	consentUp.Consented = strings.ToLower(consentUp.Consented)
	switch consentUp.Consented {

	case strings.ToLower(common.ConsentStatusAskMe):
		consentUp.Consented = common.ConsentStatusAskMe
	case strings.ToLower(common.ConsentStatusAllow):
		consentUp.Consented = common.ConsentStatusAllow
	case strings.ToLower(common.ConsentStatusDisAllow):
		consentUp.Consented = common.ConsentStatusDisAllow
	default:
		m := fmt.Sprintf("Please provide a valid value for consent; Failed to update consent: %v for org: %v user: %v", consentID, orgID, userID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	var p consent.Purpose
	p.ID = purposeID

	var con consent.Consent
	con.TemplateID = attributeID
	con.Status.Consented = consentUp.Consented
	if consentUp.Days > 0 {
		con.Status.TimeStamp = time.Now()
		con.Status.Days = consentUp.Days
	}

	/*
		1. Purpose is not in DB
		2. Purpose found, Attribute not in DB
		3. Purpose and Attribute in DB
	*/
	var found = 0
	for i := range c.Purposes {
		if c.Purposes[i].ID == purposeID {
			for j := range c.Purposes[i].Consents {
				if c.Purposes[i].Consents[j].TemplateID == attributeID {
					c.Purposes[i].Consents[j] = con
					found = 1
					break
				}
			}
			if found == 0 {
				c.Purposes[i].Consents = append(c.Purposes[i].Consents, con)
				found = 1
				break
			}
			break
		}
	}
	if found == 0 {
		p.Consents = append(p.Consents, con)
		c.Purposes = append(c.Purposes, p)
	}

	updatePurposeStatus(purposeID, o.Purposes, o.Templates, &c)

	c, err = consent.UpdatePurposes(c)
	if err != nil {
		m := fmt.Sprintf("Failed to update consent: %v for org: %v user: %v ", consentID, orgID, userID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var ch consentHistory
	ch.UserID = userID
	ch.OrgID = orgID
	ch.OrgName = o.Name
	ch.PurposeID = purposeID
	ch.PurposeName = getPurposeFromID(o.Purposes, purposeID).Name
	ch.ConsentID = c.ID.Hex()
	ch.AttributeID = attributeID
	ch.AttributeDescription = getTemplateFromOrg(o, attributeID).Consent

	ch.AttributeConsentStatus = consentUp.Consented
	if consentUp.Days > 0 {
		ch.AttributeConsentStatus = common.ConsentStatusAskMe
	}
	err = consentHistoryAttributeAdd(ch)
	if err != nil {
		m := fmt.Sprintf("Failed to update log for consent: %v for org: %v user: %v ", consentID, orgID, userID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// Trigger webhooks
	var consentedAttributes []string
	consentedAttributes = append(consentedAttributes, attributeID)
	webhookEventTypeID := webhooks.EventTypeConsentDisAllowed
	if consentUp.Consented == common.ConsentStatusAllow || consentUp.Consented == common.ConsentStatusAskMe {
		webhookEventTypeID = webhooks.EventTypeConsentAllowed
	}
	go webhooks.TriggerConsentWebhookEvent(userID, purposeID, consentID, orgID, webhooks.EventTypes[webhookEventTypeID], strconv.FormatInt(con.Status.TimeStamp.UTC().Unix(), 10), consentUp.Days, consentedAttributes)

	type consentUpdateresp struct {
		Msg    string
		Status int
	}

	updateResp := consentUpdateresp{"Consent updated successfully", http.StatusOK}
	response, _ := json.Marshal(updateResp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}
