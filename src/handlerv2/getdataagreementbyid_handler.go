package handlerv2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	"github.com/gorilla/mux"
)

// GetDataAgreementById Get a data agreement by ID
func GetDataAgreementById(w http.ResponseWriter, r *http.Request) {
	orgID := r.Header.Get(config.OrganizationId)
	purposeID := mux.Vars(r)[config.DataAgreementId]

	o, err := org.Get(orgID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", orgID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	type purposeTemplates struct {
		ID      string
		Consent string
	}

	type purposeDetails struct {
		Purpose   org.Purpose
		Templates []purposeTemplates
	}
	var pDetails purposeDetails
	for _, p := range o.Purposes {
		if p.ID == purposeID {
			pDetails.Purpose = p
		}
	}

	for _, t := range o.Templates {
		for _, pID := range t.PurposeIDs {
			if pID == purposeID {
				pDetails.Templates = append(pDetails.Templates, purposeTemplates{ID: t.ID, Consent: t.Consent})
			}
		}
	}

	response, _ := json.Marshal(pDetails)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
