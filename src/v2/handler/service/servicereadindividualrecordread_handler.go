package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/consent"
	"github.com/bb-consent/api/src/consenthistory"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/v2/token"
	"github.com/gorilla/mux"
)

type consentPurposeResp struct {
	ID            string
	ConsentID     string
	OrgID         string
	UserID        string
	Consents      consentsAndPurpose
	DataRetention DataRetentionPolicyResp
}

func ServiceReadIndividualRecordRead(w http.ResponseWriter, r *http.Request) {

	userID := token.GetUserID(r)
	purposeID := mux.Vars(r)[config.DataAgreementId]
	orgID := r.Header.Get(config.OrganizationId)

	sanitizedUserId := common.Sanitize(userID)
	sanitizedOrgId := common.Sanitize(orgID)
	sanitizedPurposeId := common.Sanitize(purposeID)

	consents, err := consent.GetByUserOrg(sanitizedUserId, sanitizedOrgId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch consents user: %v org: %v", userID, orgID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	consentID := consents.ID.Hex()

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
