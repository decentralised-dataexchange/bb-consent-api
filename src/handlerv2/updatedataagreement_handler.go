package handlerv2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	"github.com/gorilla/mux"
)

type organization struct {
	Organization org.Organization
}

// UpdateDataAgreementD Update the given data agreement by ID
func UpdateDataAgreement(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Header.Get(config.OrganizationId)
	purposeID := mux.Vars(r)[config.DataAgreementId]

	var uReq purpose
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &uReq)

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// Proceed if lawful basis of processing provided is valid
	if !isValidLawfulBasisOfProcessing(uReq.LawfulBasisOfProcessing) {
		m := fmt.Sprintf("Invalid lawful basis of processing provided")
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	tempLawfulUsage := getLawfulUsageByLawfulBasis(uReq.LawfulBasisOfProcessing)

	var found = false
	for i := range o.Purposes {
		if o.Purposes[i].ID == purposeID {
			found = true
			o.Purposes[i].Name = strings.TrimSpace(uReq.Name)
			o.Purposes[i].Description = strings.TrimSpace(uReq.Description)
			o.Purposes[i].PolicyURL = strings.TrimSpace(uReq.PolicyURL)
			o.Purposes[i].LawfulUsage = tempLawfulUsage
			o.Purposes[i].LawfulBasisOfProcessing = uReq.LawfulBasisOfProcessing
			o.Purposes[i].Jurisdiction = uReq.Jurisdiction
			o.Purposes[i].Disclosure = uReq.Disclosure
			o.Purposes[i].IndustryScope = uReq.IndustryScope
			o.Purposes[i].DataRetention = uReq.DataRetention
			o.Purposes[i].Restriction = uReq.Restriction
			o.Purposes[i].Shared3PP = uReq.Shared3PP
			if (o.Purposes[i].AttributeType != uReq.AttributeType) ||
				(o.Purposes[i].SSIID != uReq.SSIID) {
				m := fmt.Sprintf("Can not modify attributeType or SSIID for purpose: %v organization: %v",
					organizationID, purposeID)
				common.HandleError(w, http.StatusBadRequest, m, err)
				return
			}
		}
	}

	if !found {
		m := fmt.Sprintf("Failed to find purpose with ID: %v in organization: %v", purposeID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	o, err = org.Update(o)
	if err != nil {
		m := fmt.Sprintf("Failed to update purpose: %v in organization: %v", purposeID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	response, _ := json.Marshal(organization{o})
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
