package handlerv2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/orgtype"
)

type globalPolicyConfigurationResp struct {
	PolicyURL     string
	DataRetention org.DataRetention
	Jurisdiction  string
	Disclosure    string
	Type          orgtype.OrgType
	Restriction   string
	Shared3PP     bool
}

// GetGlobalPolicyConfiguration Handler to get global policy configurations
func GetGlobalPolicyConfiguration(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Header.Get(config.OrganizationId)

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// Constructing the response
	var resp globalPolicyConfigurationResp

	resp.PolicyURL = o.PolicyURL
	resp.DataRetention = o.DataRetention

	if len(strings.TrimSpace(o.Jurisdiction)) == 0 {
		resp.Jurisdiction = o.Location
		o.Jurisdiction = o.Location
	} else {
		resp.Jurisdiction = o.Jurisdiction
	}

	if len(strings.TrimSpace(o.Disclosure)) == 0 {
		resp.Disclosure = "false"
		o.Disclosure = "false"
	} else {
		resp.Disclosure = o.Disclosure
	}

	resp.Type = o.Type
	resp.Restriction = o.Restriction
	resp.Shared3PP = o.Shared3PP

	// Updating global configuration policy with defaults
	_, err = org.Update(o)
	if err != nil {
		m := fmt.Sprintf("Failed to update global configuration with defaults to organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
