package service

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

type getPolicyResp struct {
	Policy   policy.Policy `json:"policy"`
	Revision interface{}   `json:"revision"`
}

// ServiceReadPolicy
func ServiceReadPolicy(w http.ResponseWriter, r *http.Request) {
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	policyId := mux.Vars(r)[config.PolicyId]
	policyId = common.Sanitize(policyId)

	// Parse the URL query parameters
	revisionId := r.URL.Query().Get("revisionId")
	revisionId = common.Sanitize(revisionId)

	// Repository
	policyRepo := policy.PolicyRepository{}
	policyRepo.Init(organisationId)

	p, err := policyRepo.Get(policyId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch policy: %v", policyId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	var revisionResp revision.Revision

	if revisionId != "" {

		revisionResp, err = revision.GetByRevisionId(revisionId)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch revision: %v", policyId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

	} else {
		revisionResp, err = revision.GetLatestByPolicyId(policyId)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch revision: %v", policyId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
	}

	// Constructing the response
	var resp getPolicyResp
	resp.Policy = p

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(revisionResp)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
