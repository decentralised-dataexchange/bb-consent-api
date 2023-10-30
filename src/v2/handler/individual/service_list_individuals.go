package individual

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/individual"
	"github.com/bb-consent/api/src/v2/paginate"
)

type listServiceIndividualsResp struct {
	Individuals interface{}         `json:"individuals"`
	Pagination  paginate.Pagination `json:"pagination"`
}

// ServiceListIndividuals
func ServiceListIndividuals(w http.ResponseWriter, r *http.Request) {
	fmt.Println()

	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

	var resp listServiceIndividualsResp

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
			resp = listServiceIndividualsResp{
				Individuals: emptyIndividuals,
				Pagination:  result.Pagination,
			}
			common.ReturnHTTPResponse(resp, w)
			return
		}
		m := "Failed to paginate data attribute"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return

	}
	resp = listServiceIndividualsResp{
		Individuals: result.Items,
		Pagination:  result.Pagination,
	}
	common.ReturnHTTPResponse(resp, w)
}
