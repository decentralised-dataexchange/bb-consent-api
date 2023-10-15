package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	"github.com/gorilla/mux"
)

func deletePurposeIDFromTemplate(purposeID string, orgID string, templates []org.Template) error {
	for _, t := range templates {
		for _, p := range t.PurposeIDs {
			if p == purposeID {
				var template org.Template
				template.Consent = t.Consent
				template.ID = t.ID
				for _, p := range t.PurposeIDs {
					if p != purposeID {
						template.PurposeIDs = append(template.PurposeIDs, p)
					}
				}
				_, err := org.DeleteTemplates(orgID, t)
				if err != nil {
					fmt.Printf("Failed to delete template: %v from organization: %v", t.ID, orgID)
					return err
				}
				if len(template.PurposeIDs) == 0 {
					continue
				}
				err = org.AddTemplates(orgID, template)
				if err != nil {
					fmt.Printf("Failed to add template: %v from organization: %v", t.ID, orgID)
					return err
				}
				continue
			}
		}
	}
	return nil
}

// DeleteDataAgreement Deletes the given data agreement by ID
func DeleteDataAgreement(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Header.Get(config.OrganizationId)
	purposeID := mux.Vars(r)[config.DataAgreementId]

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var purposeToDelete org.Purpose
	for _, p := range o.Purposes {
		if p.ID == purposeID {
			purposeToDelete = p
		}
	}

	if purposeToDelete == (org.Purpose{}) {
		m := fmt.Sprintf("Failed to find purpose with ID: %v in organization: %v", purposeID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	//TODO: Before we delete purpose, need to remove the purpose from the templates
	err = deletePurposeIDFromTemplate(purposeID, o.ID.Hex(), o.Templates)
	if err != nil {
		m := fmt.Sprintf("Failed to update template for organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	orgResp, err := org.DeletePurposes(o.ID.Hex(), purposeToDelete)
	if err != nil {
		m := fmt.Sprintf("Failed to delete purpose: %v from organization: %v", purposeID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	response, _ := json.Marshal(organization{orgResp})
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
