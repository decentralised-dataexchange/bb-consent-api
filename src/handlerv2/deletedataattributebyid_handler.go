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

// DeleteDataAttributeById Deletes an organization data attribute
func DeleteDataAttributeById(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Header.Get(config.OrganizationId)
	templateID := mux.Vars(r)[config.DataAttributeId]

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var templateToDelete org.Template
	for _, t := range o.Templates {
		if t.ID == templateID {
			templateToDelete = t
		}
	}

	if templateToDelete.ID != templateID {
		m := fmt.Sprintf("Failed to find template with ID: %v in organization: %v", templateID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	orgResp, err := org.DeleteTemplates(o.ID.Hex(), templateToDelete)
	if err != nil {
		m := fmt.Sprintf("Failed to delete template: %v from organization: %v", templateID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	response, _ := json.Marshal(organization{orgResp})
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
