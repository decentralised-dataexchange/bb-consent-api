package dataagreement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/dataagreement"
	"github.com/bb-consent/api/src/v2/paginate"
	"github.com/bb-consent/api/src/v2/revision"
)

// ListDataAgreementsError is an error enumeration for list data agreement API.
type ListDataAgreementsError int

const (
	// ErrRevisionIDIsMissing indicates that the revisionId query param is missing.
	RevisionIDIsMissingError ListDataAgreementsError = iota
)

// Error returns the string representation of the error.
func (e ListDataAgreementsError) Error() string {
	switch e {
	case RevisionIDIsMissingError:
		return "Query param revisionId is missing!"
	default:
		return "Unknown error!"
	}
}

// ParseListDataAgreementsQueryParams parses query params for listing data agreements.
func ParseListDataAgreementsQueryParams(r *http.Request) (revisionId string, err error) {
	query := r.URL.Query()

	// Check if revisionId query param is provided.
	if r, ok := query["revisionId"]; ok && len(r) > 0 {
		return r[0], nil
	}

	return "", RevisionIDIsMissingError
}

type listDataAgreementsResp struct {
	DataAgreements interface{}         `json:"dataAgreements"`
	Pagination     paginate.Pagination `json:"pagination"`
}

func returnHTTPResponse(resp interface{}, w http.ResponseWriter) {
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// ConfigListDataAgreements
func ConfigListDataAgreements(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	var resp listDataAgreementsResp

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)
	revisionId, err := ParseListDataAgreementsQueryParams(r)
	revisionId = common.Sanitize(revisionId)
	if err != nil && errors.Is(err, RevisionIDIsMissingError) {

		darepo := dataagreement.DataAgreementRepository{}
		darepo.Init(organisationId)
		// Return all data agreements
		var dataAgreements []dataagreement.DataAgreement
		query := paginate.PaginateDBObjectsQuery{
			Filter:     darepo.DefaultFilter,
			Collection: dataagreement.Collection(),
			Context:    context.Background(),
			Limit:      limit,
			Offset:     offset,
		}
		result, err := paginate.PaginateDBObjects(query, &dataAgreements)
		if err != nil {
			if errors.Is(err, paginate.EmptyDBError) {
				emptyDataAgreements := make([]interface{}, 0)
				resp = listDataAgreementsResp{
					DataAgreements: emptyDataAgreements,
					Pagination:     result.Pagination,
				}
				returnHTTPResponse(resp, w)
				return
			}
			m := "Failed to paginate data agreement"
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return

		}
		resp = listDataAgreementsResp{
			DataAgreements: result.Items,
			Pagination:     result.Pagination,
		}
		returnHTTPResponse(resp, w)
		return

	} else {
		// Fetch revision by id
		revisionResp, err := revision.GetByRevisionId(revisionId)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch revision by id: %v", revisionId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

		// Recreate data agreement from revision
		da, err := revision.RecreateDataAgreementFromRevision(revisionResp)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch data agreement by revision: %v", revisionId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

		interfaceSlice := make([]interface{}, 0)
		interfaceSlice = append(interfaceSlice, da)

		// Constructing the response
		resp = listDataAgreementsResp{
			DataAgreements: interfaceSlice,
			Pagination: paginate.Pagination{
				CurrentPage: 1,
				TotalItems:  1,
				TotalPages:  1,
				Limit:       1,
				HasPrevious: false,
				HasNext:     false,
			},
		}

	}

	returnHTTPResponse(resp, w)
}
