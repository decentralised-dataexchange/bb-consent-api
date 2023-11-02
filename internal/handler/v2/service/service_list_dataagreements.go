package service

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/paginate"
	"github.com/bb-consent/api/internal/revision"
)

// ListDataAgreementsError is an error enumeration for list data agreement API.
type ListDataAgreementsError int

const (
	// ErrRevisionIDIsMissing indicates that the revisionId query param is missing.
	RevisionIDIsMissingError ListDataAgreementsError = iota
	LifecycleIsMissingError
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

func activeDataAgreementsFromObjectData(organisationId string) ([]interface{}, error) {
	var activeDataAgreements []interface{}
	dataAgreements, err := dataagreement.GetAllDataAgreementsWithLatestRevisionsObjectData(organisationId)
	if err != nil {
		return activeDataAgreements, err
	}

	for _, dataAgreement := range dataAgreements {
		// Recreate data agreement from revision
		activeDataAgreement, err := revision.RecreateDataAgreementFromObjectData(dataAgreement.ObjectData)
		if err != nil {
			return activeDataAgreements, err
		}
		activeDataAgreements = append(activeDataAgreements, activeDataAgreement)
	}

	return activeDataAgreements, nil
}

type listDataAgreementsResp struct {
	DataAgreements interface{}         `json:"dataAgreements"`
	Pagination     paginate.Pagination `json:"pagination"`
}

// ServiceListDataAgreements
func ServiceListDataAgreements(w http.ResponseWriter, r *http.Request) {
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

		activeDataAgreements, err := activeDataAgreementsFromObjectData(organisationId)
		if err != nil {
			common.HandleErrorV2(w, http.StatusInternalServerError, "Failed to fetch active data agreements", err)
			return
		}

		query := paginate.PaginateObjectsQuery{
			Limit:  limit,
			Offset: offset,
		}
		result := paginate.PaginateObjects(query, activeDataAgreements)

		resp = listDataAgreementsResp{
			DataAgreements: result.Items,
			Pagination:     result.Pagination,
		}
		common.ReturnHTTPResponse(resp, w)
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

	common.ReturnHTTPResponse(resp, w)
}
