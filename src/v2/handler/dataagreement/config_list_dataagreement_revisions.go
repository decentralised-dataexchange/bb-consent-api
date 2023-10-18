package dataagreement

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/dataagreement"
	"github.com/bb-consent/api/src/v2/paginate"
	"github.com/bb-consent/api/src/v2/revision"
	"github.com/gorilla/mux"
)

type listRevisionsResp struct {
	DataAgreement dataagreement.DataAgreement `json:"dataAgreement"`
	Revisions     interface{}                 `json:"revisions"`
	Pagination    paginate.Pagination         `json:"pagination"`
}

func revisionsToInterfaceSlice(revisions []revision.RevisionForHTTPResponse) []interface{} {
	interfaceSlice := make([]interface{}, len(revisions))
	for i, r := range revisions {
		interfaceSlice[i] = r
	}
	return interfaceSlice
}

// ConfigListDataAgreementRevisions
func ConfigListDataAgreementRevisions(w http.ResponseWriter, r *http.Request) {

	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)
	dataAgreementId := mux.Vars(r)[config.DataAgreementId]
	dataAgreementId = common.Sanitize(dataAgreementId)

	// Repository
	daRepo := dataagreement.DataAgreementRepository{}
	daRepo.Init(organisationId)

	da, err := daRepo.Get(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	revisions, err := revision.ListAllByDataAgreementId(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch revision: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

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
		DataAgreement: da,
		Revisions:     result.Items,
		Pagination:    result.Pagination,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
