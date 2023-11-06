package audit

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/paginate"
	"go.mongodb.org/mongo-driver/bson"
)

type listDataAgreementsResp struct {
	DataAgreements interface{}         `json:"dataAgreements"`
	Pagination     paginate.Pagination `json:"pagination"`
}

// AuditListDataAgreements
func AuditListDataAgreements(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

	darepo := dataagreement.DataAgreementRepository{}
	darepo.Init(organisationId)

	var resp listDataAgreementsResp

	var pipeline []bson.M
	pipeline = append(pipeline, bson.M{"$match": bson.M{"organisationid": organisationId, "isdeleted": false}})
	pipeline = append(pipeline, bson.M{"$sort": bson.M{"timestamp": -1}})

	// Return all data agreements
	var dataAgreements []dataagreement.DataAgreement
	query := paginate.PaginateDBObjectsQueryUsingPipeline{
		Pipeline:   pipeline,
		Collection: dataagreement.Collection(),
		Context:    context.Background(),
		Limit:      limit,
		Offset:     offset,
	}
	result, err := paginate.PaginateDBObjectsUsingPipeline(query, &dataAgreements)
	if err != nil {
		if errors.Is(err, paginate.EmptyDBError) {
			emptyDataAgreements := make([]interface{}, 0)
			resp = listDataAgreementsResp{
				DataAgreements: emptyDataAgreements,
				Pagination:     result.Pagination,
			}
			common.ReturnHTTPResponse(resp, w)
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
	common.ReturnHTTPResponse(resp, w)
}
