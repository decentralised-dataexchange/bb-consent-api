package handler

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

// UpdateDataAttributeById Updates an organization data attribute
func UpdateDataAttributeById(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Header.Get(config.OrganizationId)
	templateID := mux.Vars(r)[config.DataAttributeId]

	var uReq template
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &uReq)

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// validating PurposeIDs provided
	if uReq.PurposeIDs != nil {

		for _, p := range uReq.PurposeIDs {
			_, err := org.GetPurpose(organizationID, p)
			if err != nil {
				m := fmt.Sprintf("Invalid purposeID:%v provided;Failed to update template to organization: %v", p, o.Name)
				common.HandleError(w, http.StatusBadRequest, m, err)
				return
			}
		}
	}

	var found = false

	for i := range o.Templates {
		if o.Templates[i].ID == templateID {
			found = true

			// for partial updation
			if strings.TrimSpace(uReq.Consent) != "" {
				o.Templates[i].Consent = uReq.Consent
			}
			// only updating if any valid purpose id was given
			if uReq.PurposeIDs != nil {
				o.Templates[i].PurposeIDs = uReq.PurposeIDs
			}
		}
	}

	if !found {
		m := fmt.Sprintf("Failed to find template with ID: %v in organization: %v", templateID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	o, err = org.Update(o)
	if err != nil {
		m := fmt.Sprintf("Failed to update template: %v in organization: %v", templateID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	response, _ := json.Marshal(organization{o})

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
