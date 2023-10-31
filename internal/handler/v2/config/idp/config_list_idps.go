package idp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/idp"
	"github.com/bb-consent/api/internal/paginate"
)

func returnHTTPResponse(resp interface{}, w http.ResponseWriter) {
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

type listIdpsResp struct {
	Idps       interface{}         `json:"idps"`
	Pagination paginate.Pagination `json:"pagination"`
}

// ConfigListIdps
func ConfigListIdps(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	var resp listIdpsResp

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)

	// Repository
	idpRepo := idp.IdentityProviderRepository{}
	idpRepo.Init(organisationId)

	// Paginate idps
	var idps []idp.IdentityProvider
	query := paginate.PaginateDBObjectsQuery{
		Filter:     idpRepo.DefaultFilter,
		Collection: idp.Collection(),
		Context:    context.Background(),
		Limit:      limit,
		Offset:     offset,
	}
	result, err := paginate.PaginateDBObjects(query, &idps)
	if err != nil {
		if errors.Is(err, paginate.EmptyDBError) {
			emptyIdps := make([]interface{}, 0)
			resp = listIdpsResp{
				Idps:       emptyIdps,
				Pagination: result.Pagination,
			}
			returnHTTPResponse(resp, w)
			return
		}
		m := "Failed to paginate idps"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return

	}
	resp = listIdpsResp{
		Idps:       result.Items,
		Pagination: result.Pagination,
	}

	returnHTTPResponse(resp, w)
}
