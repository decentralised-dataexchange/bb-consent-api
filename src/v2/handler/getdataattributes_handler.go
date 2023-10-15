package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
)

type getTemplatesResp struct {
	OrgID     string
	Templates []org.Template
}

// GetDataAttributes Gets an organization data attributes
func GetDataAttributes(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Header.Get(config.OrganizationId)

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(getTemplatesResp{OrgID: o.ID.Hex(), Templates: o.Templates})
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
