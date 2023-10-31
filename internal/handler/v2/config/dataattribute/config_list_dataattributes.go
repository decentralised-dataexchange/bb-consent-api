package dataattribute

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/paginate"
	"github.com/bb-consent/api/internal/revision"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func dataAttributesForList(res []dataagreement.DataAgreement) []dataAttributeForLists {

	var dataAttributes []dataAttributeForLists
	for i := range res {
		for _, dA := range res[i].DataAttributes {
			var dataAttribute dataAttributeForLists
			dataAttribute.Id = dA.Id
			dataAttribute.Name = dA.Name
			dataAttribute.Description = dA.Description
			dataAttribute.Sensitivity = dA.Sensitivity
			dataAttribute.Category = dA.Category
			dataAttribute.DataAgreement.Id = res[i].Id.Hex()
			dataAttribute.DataAgreement.Purpose = res[i].Purpose
			dataAttributes = append(dataAttributes, dataAttribute)

		}

	}

	return dataAttributes
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

func dataAttributesToInterfaceSlice(dataAttributes []dataAttributeForLists) []interface{} {
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

type dataAgreementForDataAttribute struct {
	Id      string `json:"id" bson:"_id,omitempty"`
	Purpose string `json:"purpose"`
}

type dataAttributeForLists struct {
	Id            primitive.ObjectID            `json:"id" bson:"_id,omitempty"`
	Name          string                        `json:"name" valid:"required"`
	Description   string                        `json:"description" valid:"required"`
	Sensitivity   bool                          `json:"sensitivity"`
	Category      string                        `json:"category"`
	DataAgreement dataAgreementForDataAttribute `json:"dataAgreement"`
}

type listDataAttributesResp struct {
	DataAttributes interface{}         `json:"attributes"`
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

	// Repository
	darepo := dataagreement.DataAgreementRepository{}
	darepo.Init(organisationId)

	if err != nil && errors.Is(err, RevisionIDIsMissingError) {

		methodOfUse, err := ParseMethodOfUseDataAttributesQueryParams(r)
		methodOfUse = common.Sanitize(methodOfUse)
		var res []dataagreement.DataAgreement

		if err != nil && errors.Is(err, MethodOfUseIsMissingError) {

			// Return all data attributes
			res, err = darepo.GetAll()
			if err != nil {
				m := fmt.Sprintf("Failed to fetch data attribute by method of use: %v", methodOfUse)
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}

		} else {
			// List by method of use
			res, err = darepo.GetByMethodOfUse(methodOfUse)
			if err != nil {
				m := fmt.Sprintf("Failed to fetch data attribute by method of use: %v", methodOfUse)
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}

		}
		dataAttributes := dataAttributesForList(res)

		query := paginate.PaginateObjectsQuery{
			Limit:  limit,
			Offset: offset,
		}
		interfaceSlice := dataAttributesToInterfaceSlice(dataAttributes)
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

		// Recreate data agreement from revision
		da, err := revision.RecreateDataAgreementFromRevision(revisionResp)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch data attribute by revision: %v", revisionId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

		var dAttributes []dataAttributeForLists
		for _, a := range da.DataAttributes {
			var dA dataAttributeForLists
			dA.Id = a.Id
			dA.Name = a.Name
			dA.Description = a.Description
			dA.Sensitivity = a.Sensitivity
			dA.Category = a.Category
			dA.DataAgreement.Id = da.Id.Hex()
			dA.DataAgreement.Purpose = da.Purpose
			dAttributes = append(dAttributes, dA)
		}

		interfaceSlice := make([]interface{}, 0)
		interfaceSlice = append(interfaceSlice, dAttributes)

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
