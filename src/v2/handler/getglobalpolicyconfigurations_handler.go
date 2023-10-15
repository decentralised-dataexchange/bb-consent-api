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

type getPolicyResp struct {
	Policy   policy.Policy `json:"policy"`
	Revision interface{}   `json:"revision"`
}

// GetGlobalPolicyConfiguration Handler to get global policy configurations
func GetGlobalPolicyConfiguration(w http.ResponseWriter, r *http.Request) {
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	policyId := mux.Vars(r)[config.PolicyId]

	// Parse the URL query parameters
	queryParams := r.URL.Query()
	revisionId := queryParams.Get("revisionId")

	p, err := policy.Get(policyId, organisationId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch policy: %v", policyId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	var revision policy.Revision
	if revisionId != "" {
		for _, p := range p.Revisions {
			if p.Id == revisionId {
				revision = p
			}
		}

	} else {
		revision = p.Revisions[len(p.Revisions)-1]
	}

	// Constructing the response
	var resp getPolicyResp
	resp.Policy = p

	var revisionForHTTPResponse policy.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(revision)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
