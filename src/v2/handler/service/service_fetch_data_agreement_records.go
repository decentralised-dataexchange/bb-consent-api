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
)

type fetchDataAgreementRecordsResp struct {
	DataAgreementRecords interface{}         `json:"dataAgreementRecords"`
	Pagination           paginate.Pagination `json:"pagination"`
}

func ServiceFetchDataAgreementRecords(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	// Return all data agreement records
	var dataAgreementRecords []daRecord.DataAgreementRecord
	query := paginate.PaginateDBObjectsQuery{
		Filter:     darRepo.DefaultFilter,
		Collection: daRecord.Collection(),
		Context:    context.Background(),
		Limit:      limit,
		Offset:     offset,
	}
	var resp fetchDataAgreementRecordsResp
	result, err := paginate.PaginateDBObjects(query, &dataAgreementRecords)
	if err != nil {
		if errors.Is(err, paginate.EmptyDBError) {
			emptyDataAgreementRecords := make([]interface{}, 0)
			resp = fetchDataAgreementRecordsResp{
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
	resp = fetchDataAgreementRecordsResp{
		DataAgreementRecords: result.Items,
		Pagination:           result.Pagination,
	}
	common.ReturnHTTPResponse(resp, w)
}
