package policy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/paginate"
	"github.com/bb-consent/api/internal/policy"
	"github.com/bb-consent/api/internal/revision"
	"go.mongodb.org/mongo-driver/bson"
)

// ListPoliciesError is an error enumeration for list policies API.
type ListPoliciesError int

const (
	// ErrRevisionIDIsMissing indicates that the revisionId query param is missing.
	RevisionIDIsMissingError ListPoliciesError = iota
)

// Error returns the string representation of the error.
func (e ListPoliciesError) Error() string {
	switch e {
	case RevisionIDIsMissingError:
		return "Query param revisionId is missing!"
	default:
		return "Unknown error!"
	}
}

// ParseListPoliciesQueryParams parses query params for listing policies.
func ParseListPoliciesQueryParams(r *http.Request) (revisionId string, err error) {
	query := r.URL.Query()

	// Check if revisionId query param is provided.
	if r, ok := query["revisionId"]; ok && len(r) > 0 {
		return r[0], nil
	}

	return "", RevisionIDIsMissingError
}

type listPoliciesResp struct {
	Policies   interface{}         `json:"policies"`
	Pagination paginate.Pagination `json:"pagination"`
}

func returnHTTPResponse(resp interface{}, w http.ResponseWriter) {
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// ConfigListPolicies
func ConfigListPolicies(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	var resp listPoliciesResp

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)
	revisionId, err := ParseListPoliciesQueryParams(r)
	revisionId = common.Sanitize(revisionId)
	if err != nil && errors.Is(err, RevisionIDIsMissingError) {
		// Repository
		prepo := policy.PolicyRepository{}
		prepo.Init(organisationId)

		pipeline, err := policy.CreatePipelineForFilteringPolicies(organisationId)
		if err != nil {
			m := "Failed to create pipeline"
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
		// Return all policies
		var policies []policy.Policy
		pipeline = append(pipeline, bson.M{"$sort": bson.M{"timestamp": -1}})
		query := paginate.PaginateDBObjectsQueryUsingPipeline{
			Pipeline:   pipeline,
			Collection: policy.Collection(),
			Context:    context.Background(),
			Limit:      limit,
			Offset:     offset,
		}
		result, err := paginate.PaginateDBObjectsUsingPipeline(query, &policies)
		if err != nil {
			if errors.Is(err, paginate.EmptyDBError) {
				emptyPolicies := make([]interface{}, 0)
				resp = listPoliciesResp{
					Policies:   emptyPolicies,
					Pagination: result.Pagination,
				}
				returnHTTPResponse(resp, w)
				return
			}
			m := "Failed to paginate policy"
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return

		}
		resp = listPoliciesResp{
			Policies:   result.Items,
			Pagination: result.Pagination,
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

		// Recreate policy from revision
		p, err := revision.RecreatePolicyFromRevision(revisionResp)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch policy by revision: %v", revisionId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

		interfaceSlice := make([]interface{}, 0)
		interfaceSlice = append(interfaceSlice, p)

		// Constructing the response
		resp = listPoliciesResp{
			Policies: interfaceSlice,
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
