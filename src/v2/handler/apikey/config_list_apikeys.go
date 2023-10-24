package apikey

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/apikey"
	"github.com/bb-consent/api/src/v2/paginate"
)

type listApiKeyResp struct {
	Apikeys    interface{}         `json:"apiKeys" valid:"required"`
	Pagination paginate.Pagination `json:"pagination"`
}

// ConfigListApiKey
func ConfigListApiKey(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

	// Repository
	apiKeyRepo := apikey.ApiKeyRepository{}
	apiKeyRepo.Init(organisationId)

	// Return all api keys
	var apikeys []apikey.ApiKey
	query := paginate.PaginateDBObjectsQuery{
		Filter:     apiKeyRepo.DefaultFilter,
		Collection: apikey.Collection(),
		Context:    context.Background(),
		Limit:      limit,
		Offset:     offset,
	}

	var resp listApiKeyResp
	result, err := paginate.PaginateDBObjects(query, &apikeys)
	if err != nil {
		if errors.Is(err, paginate.EmptyDBError) {
			emptyApikeys := make([]interface{}, 0)
			resp = listApiKeyResp{
				Apikeys:    emptyApikeys,
				Pagination: result.Pagination,
			}
			common.ReturnHTTPResponse(resp, w)
			return
		}
		m := "Failed to paginate api keys"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return

	}
	resp = listApiKeyResp{
		Apikeys:    result.Items,
		Pagination: result.Pagination,
	}
	common.ReturnHTTPResponse(resp, w)

}
