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
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/consent"
	"github.com/bb-consent/api/src/consenthistory"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/webhooks"
	"github.com/gorilla/mux"
)

type purposeAllUpReq struct {
	Consented string `valid:"required"`
}

type consentsAndPurpose struct {
	Purpose  org.Purpose
	Count    ConsentCount
	Consents []consentResp
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

type consentStatus struct {
	Consented string
	TimeStamp time.Time
	Days      int
	Remaining int
}

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

// ConsentsResp Consent response struct definition
type ConsentsResp struct {
	ID                  string
	OrgID               string
	UserID              string
	ConsentsAndPurposes []consentsAndPurpose
}

func getPurposeFromID(p []org.Purpose, purposeID string) org.Purpose {
	for _, e := range p {
		if e.ID == purposeID {
			return e
		}
	}
	return org.Purpose{}
}

func createConsentGetResponse(c consent.Consents, o org.Organization) ConsentsResp {
	var cResp ConsentsResp
	cResp.ID = c.ID.Hex()
	cResp.OrgID = c.OrgID
	cResp.UserID = c.UserID

	for _, p := range o.Purposes {
		// Filtering templates corresponding to the purpose ID
		templatesWithPurpose := getTemplateswithPurpose(p.ID, o.Templates)

		// Filtering consents corresponding to purpose ID
		cons := getConsentsWithPurpose(p.ID, c)

		conResp := createConsentResponse(templatesWithPurpose, cons, p)

		cResp.ConsentsAndPurposes = append(cResp.ConsentsAndPurposes, conResp)
	}
	return cResp
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

func getTemplateswithPurpose(purposeID string, templates []org.Template) []org.Template {
	var t []org.Template
	for _, template := range templates {
		for _, pID := range template.PurposeIDs {
			if pID == purposeID {
				t = append(t, template)
				break
			}
		}
	}
	return t
}

func getConsentsWithPurpose(purposeID string, c consent.Consents) []consent.Consent {
	for _, p := range c.Purposes {
		if p.ID == purposeID {
			return p.Consents
		}
	}
	return []consent.Consent{}
}

func createConsentResponse(templates []org.Template, consents []consent.Consent, purpose org.Purpose) consentsAndPurpose {
	var cp consentsAndPurpose
	for _, template := range templates {
		var conResp consentResp
		conResp.ID = template.ID
		conResp.Description = template.Consent

		// Fetching consents matching a Template ID
		c := getConsentWithTemplateID(template.ID, consents)

		if (consent.Consent{}) == c {
			if purpose.LawfulUsage {
				conResp.Status.Consented = common.ConsentStatusAllow
			} else {
				conResp.Status.Consented = common.ConsentStatusDisAllow
			}
		} else {
			conResp.Status.Consented = c.Status.Consented
			conResp.Value = c.Value
			if c.Status.Days != 0 {
				conResp.Status.Days = c.Status.Days
				conResp.Status.Remaining = c.Status.Days - int((time.Now().Sub(c.Status.TimeStamp).Hours())/24)
				if conResp.Status.Remaining <= 0 {
					conResp.Status.Consented = common.ConsentStatusDisAllow
					conResp.Status.Remaining = 0
				} else {
					conResp.Status.TimeStamp = c.Status.TimeStamp
				}

			}
		}
		cp.Consents = append(cp.Consents, conResp)
	}
	cp.Purpose = purpose
	cp.Count = getConsentCount(cp)
	return cp
}

func getConsentWithTemplateID(templateID string, consents []consent.Consent) consent.Consent {
	for _, c := range consents {
		if c.TemplateID == templateID {
			return c
		}
	}
	return consent.Consent{}
}

func getConsentCount(cp consentsAndPurpose) ConsentCount {
	var c ConsentCount
	var disallowCount = 0

	for _, p := range cp.Consents {
		c.Total++
		if (p.Status.Consented == common.ConsentStatusDisAllow) || (p.Status.Consented == "DisAllow") {
			disallowCount++
		}
	}
	c.Consented = c.Total - disallowCount
	return c
}

func ServiceCreateIndividualConsentRecord(w http.ResponseWriter, r *http.Request) {

	userID := token.GetUserID(r)
	purposeID := mux.Vars(r)[config.DataAgreementId]
	orgID := r.Header.Get(config.OrganizationId)

	sanitizedUserId := common.Sanitize(userID)
	sanitizedOrgId := common.Sanitize(orgID)

	consents, err := consent.GetByUserOrg(sanitizedUserId, sanitizedOrgId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch consents user: %v org: %v", userID, orgID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	consentID := consents.ID.Hex()

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
