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
	"go.mongodb.org/mongo-driver/bson"
)

type vFetchIndividualDataAgreementRecordsResp struct {
	DataAgreementRecords interface{}         `json:"dataAgreementRecords"`
	Pagination           paginate.Pagination `json:"pagination"`
}

func ServiceFetchIndividualDataAgreementRecords(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))
	individualId := common.Sanitize(r.Header.Get(config.IndividualHeaderKey))

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	// Return all data agreement records
	var dataAgreementRecords []daRecord.DataAgreementRecord
	query := paginate.PaginateDBObjectsQuery{
		Filter:     common.CombineFilters(darRepo.DefaultFilter, bson.M{"individualid": individualId}),
		Collection: daRecord.Collection(),
		Context:    context.Background(),
		Limit:      limit,
		Offset:     offset,
	}
	var resp vFetchIndividualDataAgreementRecordsResp
	result, err := paginate.PaginateDBObjects(query, &dataAgreementRecords)
	if err != nil {
		if errors.Is(err, paginate.EmptyDBError) {
			emptyDataAgreementRecords := make([]interface{}, 0)
			resp = vFetchIndividualDataAgreementRecordsResp{
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
	resp = vFetchIndividualDataAgreementRecordsResp{
		DataAgreementRecords: result.Items,
		Pagination:           result.Pagination,
	}
	common.ReturnHTTPResponse(resp, w)
}
