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
	"github.com/bb-consent/api/src/orgtype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type globalPolicyConfigurationReq struct {
	PolicyURL       string
	RetentionPeriod int
	Jurisdiction    string
	Disclosure      string
	TypeID          string `valid:"required"`
	Restriction     string
	Shared3PP       bool
}

// UpdateGlobalPolicyConfiguration Handler to update global policy configuration
func UpdateGlobalPolicyConfiguration(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Header.Get(config.OrganizationId)

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var policyReq globalPolicyConfigurationReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &policyReq)

	// Update global policy configuration for the org
	o.PolicyURL = policyReq.PolicyURL

	if len(strings.TrimSpace(policyReq.Jurisdiction)) != 0 {
		o.Jurisdiction = policyReq.Jurisdiction
	}

	o.Restriction = policyReq.Restriction
	o.Shared3PP = policyReq.Shared3PP

	if policyReq.Disclosure == "false" || policyReq.Disclosure == "true" {
		o.Disclosure = policyReq.Disclosure
	}

	// Check if type id is valid bson objectid hex
	if !primitive.IsValidObjectID(policyReq.TypeID) {
		m := fmt.Sprintf("Invalid organization type ID: %v", policyReq.TypeID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	orgType, err := orgtype.Get(policyReq.TypeID)
	if err != nil {
		m := fmt.Sprintf("Invalid organization type ID: %v", policyReq.TypeID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	o.Type = orgType

	if policyReq.RetentionPeriod > 0 {
		o.DataRetention.RetentionPeriod = int64(policyReq.RetentionPeriod)
		o.DataRetention.Enabled = true
	} else {
		o.DataRetention.RetentionPeriod = 0
		o.DataRetention.Enabled = false
	}

	// Updating global configuration policy with defaults
	o, err = org.Update(o)
	if err != nil {
		m := fmt.Sprintf("Failed to update global configuration to organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// Constructing the response
	var resp globalPolicyConfigurationResp
	resp.PolicyURL = o.PolicyURL
	resp.DataRetention = o.DataRetention
	resp.Jurisdiction = o.Jurisdiction
	resp.Disclosure = o.Disclosure
	resp.Type = o.Type
	resp.Restriction = o.Restriction
	resp.Shared3PP = o.Shared3PP

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
