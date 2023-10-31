package policy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/paginate"
	"github.com/bb-consent/api/internal/policy"
	"github.com/bb-consent/api/internal/revision"
	"github.com/gorilla/mux"
)

type listRevisionsResp struct {
	Policy     policy.Policy       `json:"policy"`
	Revisions  interface{}         `json:"revisions"`
	Pagination paginate.Pagination `json:"pagination"`
}

func revisionsToInterfaceSlice(revisions []revision.RevisionForHTTPResponse) []interface{} {
	interfaceSlice := make([]interface{}, len(revisions))
	for i, r := range revisions {
		interfaceSlice[i] = r
	}
	return interfaceSlice
}

// ConfigListPolicyRevisions
func ConfigListPolicyRevisions(w http.ResponseWriter, r *http.Request) {

	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)
	policyId := mux.Vars(r)[config.PolicyId]
	policyId = common.Sanitize(policyId)

	// Repository
	policyRepo := policy.PolicyRepository{}
	policyRepo.Init(organisationId)

	p, err := policyRepo.Get(policyId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch policy: %v", policyId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	revisions, err := revision.ListAllByPolicyId(policyId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch revision: %v", policyId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Limit - Total number of items in current page
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 0 {
		limit = 10
	}

	// Offset - Total number of items to be skipped
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	query := paginate.PaginateObjectsQuery{
		Limit:  limit,
		Offset: offset,
	}

	revisionForHTTPResponses := make([]revision.RevisionForHTTPResponse, len(revisions))
	for i, r := range revisions {
		revisionForHTTPResponses[i].Init(r)
	}

	interfaceSlice := revisionsToInterfaceSlice(revisionForHTTPResponses)
	result := paginate.PaginateObjects(query, interfaceSlice)

	var resp = listRevisionsResp{
		Policy:     p,
		Revisions:  result.Items,
		Pagination: result.Pagination,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
