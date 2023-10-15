package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/consent"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/token"
)

func ServiceListIndividualRecordList(w http.ResponseWriter, r *http.Request) {
	userID := token.GetUserID(r)
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
