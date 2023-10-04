package handlerv2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
)

// GetOrganizationByID Gets a single organization by given id
func GetOrganizationByID(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Header.Get(config.OrganizationId)
	o, err := org.Get(organizationID)

	if err != nil {
		m := fmt.Sprintf("Failed to get organization by ID :%v", organizationID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	response, _ := json.Marshal(organization{o})
	w.Write(response)
}
