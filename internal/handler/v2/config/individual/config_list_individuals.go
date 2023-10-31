package individual

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/individual"
	"github.com/bb-consent/api/internal/paginate"
)

func returnHTTPResponse(resp interface{}, w http.ResponseWriter) {
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

type listIndividualsResp struct {
	Individuals interface{}         `json:"individuals"`
	Pagination  paginate.Pagination `json:"pagination"`
}

// ConfigListIndividuals
func ConfigListIndividuals(w http.ResponseWriter, r *http.Request) {
	fmt.Println()

	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

	var resp listIndividualsResp

	// Return all individuals
	var individuals []individual.Individual
	query := paginate.PaginateDBObjectsQuery{
		Filter:     individualRepo.DefaultFilter,
		Collection: individual.Collection(),
		Context:    context.Background(),
		Limit:      limit,
		Offset:     offset,
	}
	result, err := paginate.PaginateDBObjects(query, &individuals)
	if err != nil {
		if errors.Is(err, paginate.EmptyDBError) {
			emptyIndividuals := make([]interface{}, 0)
			resp = listIndividualsResp{
				Individuals: emptyIndividuals,
				Pagination:  result.Pagination,
			}
			returnHTTPResponse(resp, w)
			return
		}
		m := "Failed to paginate data attribute"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return

	}
	resp = listIndividualsResp{
		Individuals: result.Items,
		Pagination:  result.Pagination,
	}
	returnHTTPResponse(resp, w)
}
