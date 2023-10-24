package dataattribute

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/dataagreement"
	"github.com/bb-consent/api/src/v2/dataattribute"
	"github.com/bb-consent/api/src/v2/paginate"
	"github.com/bb-consent/api/src/v2/revision"
)

// ListDataAttributesError is an error enumeration for list data attribute API.
type ListDataAttributesError int

const (
	// ErrRevisionIDIsMissing indicates that the revisionId query param is missing.
	RevisionIDIsMissingError ListDataAttributesError = iota
	MethodOfUseIsMissingError
)

// Error returns the string representation of the error.
func (e ListDataAttributesError) Error() string {
	switch e {
	case RevisionIDIsMissingError:
		return "Query param revisionId is missing!"
	case MethodOfUseIsMissingError:
		return "Query param method of use is missing!"
	default:
		return "Unknown error!"
	}
}

// ParseListDataAttributesQueryParams parses query params for listing data attributes.
func ParseListDataAttributesQueryParams(r *http.Request) (revisionId string, err error) {
	query := r.URL.Query()

	// Check if revisionId query param is provided.
	if r, ok := query["revisionId"]; ok && len(r) > 0 {
		return r[0], nil
	}

	return "", RevisionIDIsMissingError
}

// ParseMethodOfUseDataAttributesQueryParam parses query method of use param for listing data attributes.
func ParseMethodOfUseDataAttributesQueryParams(r *http.Request) (methodOfUse string, err error) {
	query := r.URL.Query()

	// Check if revisionId query param is provided.
	if r, ok := query["methodOfUse"]; ok && len(r) > 0 {
		return r[0], nil
	}

	return "", MethodOfUseIsMissingError
}

func dataAttributesToInterfaceSlice(dataAttributes []dataattribute.DataAttributeForLists) []interface{} {
	interfaceSlice := make([]interface{}, len(dataAttributes))
	for i, r := range dataAttributes {
		interfaceSlice[i] = r
	}
	return interfaceSlice
}

func returnHTTPResponse(resp interface{}, w http.ResponseWriter) {
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

type listDataAttributesResp struct {
	DataAttributes interface{}         `json:"dataAttributes"`
	Pagination     paginate.Pagination `json:"pagination"`
}

// ConfigListDataAttributes
func ConfigListDataAttributes(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	var resp listDataAttributesResp

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)
	revisionId, err := ParseListDataAttributesQueryParams(r)
	revisionId = common.Sanitize(revisionId)

	if err != nil && errors.Is(err, RevisionIDIsMissingError) {

		methodOfUse, err := ParseMethodOfUseDataAttributesQueryParams(r)
		methodOfUse = common.Sanitize(methodOfUse)
		var res []dataattribute.DataAttributeForLists
		if err != nil && errors.Is(err, MethodOfUseIsMissingError) {

			// Return all data attributes
			res, err = dataattribute.ListDataAttributesWithDataAgreement(organisationId)
			if err != nil {
				m := fmt.Sprintf("Failed to fetch data attribute by method of use: %v", methodOfUse)
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}

		} else {
			// List by method of use
			res, err = dataattribute.ListDataAttributesBasedOnMethodOfUse(methodOfUse, organisationId)
			if err != nil {
				m := fmt.Sprintf("Failed to fetch data attribute by method of use: %v", methodOfUse)
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}

		}
		query := paginate.PaginateObjectsQuery{
			Limit:  limit,
			Offset: offset,
		}
		interfaceSlice := dataAttributesToInterfaceSlice(res)
		result := paginate.PaginateObjects(query, interfaceSlice)
		resp = listDataAttributesResp{
			DataAttributes: result.Items,
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

		// Recreate data attribute from revision
		da, err := revision.RecreateDataAttributeFromRevision(revisionResp)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch data attribute by revision: %v", revisionId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
		// Repository
		darepo := dataagreement.DataAgreementRepository{}
		darepo.Init(organisationId)

		var dataAgreements []dataattribute.DataAgreementForDataAttribute
		for _, a := range da.AgreementIds {
			var dA dataattribute.DataAgreementForDataAttribute
			dataAgreement, err := darepo.Get(a)
			if err != nil {
				m := fmt.Sprintf("Failed to fetch data agreement by revision: %v", revisionId)
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}
			dA.Id = dataAgreement.Id.Hex()
			dA.Purpose = dataAgreement.Purpose
			dataAgreements = append(dataAgreements, dA)
		}

		var dataAttributes dataattribute.DataAttributeForLists
		dataAttributes.AgreementData = dataAgreements

		interfaceSlice := make([]interface{}, 0)
		interfaceSlice = append(interfaceSlice, da)

		// Constructing the response
		resp = listDataAttributesResp{
			DataAttributes: interfaceSlice,
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
