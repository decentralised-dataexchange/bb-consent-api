package policy

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/policy"
	"github.com/bb-consent/api/internal/revision"
	"github.com/gorilla/mux"
)

// ConfigDeletePolicy
func ConfigDeletePolicy(w http.ResponseWriter, r *http.Request) {
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)
	policyId := mux.Vars(r)[config.PolicyId]
	policyId = common.Sanitize(policyId)

	// Repository
	policyRepo := policy.PolicyRepository{}
	policyRepo.Init(organisationId)

	currentPolicy, err := policyRepo.Get(policyId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch policy: %v", policyId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	currentRevision, err := revision.GetLatestByPolicyId(policyId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch revisions: %v", policyId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	currentPolicy.IsDeleted = true

	_, err = policyRepo.Update(currentPolicy)
	if err != nil {
		m := fmt.Sprintf("Failed to delete policy: %v", policyId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(currentRevision)

	response, _ := json.Marshal(revisionForHTTPResponse)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
