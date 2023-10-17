package dataattribute

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/dataattribute"
	"github.com/bb-consent/api/src/paginate"
	"github.com/bb-consent/api/src/revision"
	"github.com/gorilla/mux"
)

type listRevisionsResp struct {
	DataAttribute dataattribute.DataAttribute `json:"dataAttributes"`
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

// ConfigListDataAttributeRevisions
func ConfigListDataAttributeRevisions(w http.ResponseWriter, r *http.Request) {

	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)
	dataAttributeId := mux.Vars(r)[config.DataAttributeId]
	dataAttributeId = common.Sanitize(dataAttributeId)

	// Repository
	dataAttributeRepo := dataattribute.DataAttributeRepository{}
	dataAttributeRepo.Init(organisationId)

	da, err := dataAttributeRepo.Get(dataAttributeId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data attribute: %v", dataAttributeId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	revisions, err := revision.ListAllByDataAttributeId(dataAttributeId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch revision: %v", dataAttributeId)
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
		DataAttribute: da,
		Revisions:     result.Items,
		Pagination:    result.Pagination,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
