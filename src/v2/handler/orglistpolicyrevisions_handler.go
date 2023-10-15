package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/paginate"
	"github.com/bb-consent/api/src/policy"
	"github.com/gorilla/mux"
)

type listRevisionsResp struct {
	Policy     policy.Policy       `json:"policy"`
	Revisions  interface{}         `json:"revisions"`
	Pagination paginate.Pagination `json:"pagination"`
}

func revisionsToInterfaceSlice(revisions []policy.RevisionForHTTPResponse) []interface{} {
	interfaceSlice := make([]interface{}, len(revisions))
	for i, r := range revisions {
		interfaceSlice[i] = r
	}
	return interfaceSlice
}

// OrgListPolicyRevisions Handler to list global policy revisions
func OrgListPolicyRevisions(w http.ResponseWriter, r *http.Request) {

	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)
	policyId := mux.Vars(r)[config.PolicyId]

	p, err := policy.Get(policyId, organisationId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch policy: %v", policyId)
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

	revisionForHTTPResponses := make([]policy.RevisionForHTTPResponse, len(p.Revisions))
	for i, r := range p.Revisions {
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
