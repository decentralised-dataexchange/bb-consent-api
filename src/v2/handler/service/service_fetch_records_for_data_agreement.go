package service

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	daRecord "github.com/bb-consent/api/src/v2/dataagreement_record"
	"github.com/bb-consent/api/src/v2/paginate"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

type fetchRecordsForDataAgreementResp struct {
	DataAgreementRecords interface{}         `json:"dataAgreementRecords"`
	Pagination           paginate.Pagination `json:"pagination"`
}

func ServiceFetchRecordsForDataAgreement(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))
	dataAgreementId := common.Sanitize(mux.Vars(r)[config.DataAgreementId])

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	// Return all data agreement records
	var dataAgreementRecords []daRecord.DataAgreementRecord
	query := paginate.PaginateDBObjectsQuery{
		Filter:     common.CombineFilters(darRepo.DefaultFilter, bson.M{"dataagreementid": dataAgreementId}),
		Collection: daRecord.Collection(),
		Context:    context.Background(),
		Limit:      limit,
		Offset:     offset,
	}
	var resp fetchRecordsForDataAgreementResp
	result, err := paginate.PaginateDBObjects(query, &dataAgreementRecords)
	if err != nil {
		if errors.Is(err, paginate.EmptyDBError) {
			emptyDataAgreementRecords := make([]interface{}, 0)
			resp = fetchRecordsForDataAgreementResp{
				DataAgreementRecords: emptyDataAgreementRecords,
				Pagination:           result.Pagination,
			}
			common.ReturnHTTPResponse(resp, w)
			return
		}
		m := "Failed to paginate data agreement records"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return

	}
	resp = fetchRecordsForDataAgreementResp{
		DataAgreementRecords: result.Items,
		Pagination:           result.Pagination,
	}
	common.ReturnHTTPResponse(resp, w)
}
