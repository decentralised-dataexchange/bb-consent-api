package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/policy"
	"github.com/gorilla/mux"
)

// OrgDeletePolicy Handler to delete global policy revision
func OrgDeletePolicy(w http.ResponseWriter, r *http.Request) {
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)
	policyId := mux.Vars(r)[config.PolicyId]

	currentPolicy, err := policy.Get(policyId, organisationId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch policy: %v", policyId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	currentRevision := currentPolicy.Revisions[len(currentPolicy.Revisions)-1]

	currentPolicy.IsDeleted = true

	_, err = policy.Update(currentPolicy, organisationId)
	if err != nil {
		m := fmt.Sprintf("Failed to delete policy: %v", policyId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	var revisionForHTTPResponse policy.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(currentRevision)

	response, _ := json.Marshal(revisionForHTTPResponse)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
