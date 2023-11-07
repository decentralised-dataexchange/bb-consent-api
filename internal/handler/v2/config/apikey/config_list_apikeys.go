package apikey

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/bb-consent/api/internal/apikey"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/paginate"
	"go.mongodb.org/mongo-driver/bson"
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

	var pipeline []bson.M
	pipeline = append(pipeline, bson.M{"$match": bson.M{"organisationid": organisationId, "isdeleted": false}})
	pipeline = append(pipeline, bson.M{"$sort": bson.M{"timestamp": -1}})

	query := paginate.PaginateDBObjectsQueryUsingPipeline{
		Pipeline:   pipeline,
		Collection: apikey.Collection(),
		Context:    context.Background(),
		Limit:      limit,
		Offset:     offset,
	}

	var resp listApiKeyResp
	result, err := paginate.PaginateDBObjectsUsingPipeline(query, &apikeys)
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
