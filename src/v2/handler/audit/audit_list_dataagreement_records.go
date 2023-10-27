package audit

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

type DataAgreementForListDataAgreementRecord struct {
	Id      string `json:"id" bson:"_id,omitempty"`
	Purpose string `json:"purpose"`
}

type fetchDataAgreementRecordsResp struct {
	DataAgreementRecords interface{}         `json:"dataAgreementRecords"`
	Pagination           paginate.Pagination `json:"pagination"`
}

// AuditListDataAgreementRecords
func AuditListDataAgreementRecords(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	id, err := daRecord.ParseQueryParams(r, config.Id, daRecord.IdIsMissingError)
	if err != nil && errors.Is(err, daRecord.IdIsMissingError) {
		log.Println(err)
	}

	lawfulBasis, err := daRecord.ParseQueryParams(r, config.LawfulBasis, daRecord.LawfulBasisIsMissingError)
	if err != nil && errors.Is(err, daRecord.LawfulBasisIsMissingError) {
		log.Println(err)
	}

	pipeline, err := daRecord.CreatePipelineForFilteringDataAgreementRecords(organisationId, id, lawfulBasis)
	if err != nil {
		m := "Failed to create pipeline"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	pipeline = append(pipeline, bson.M{"$sort": bson.M{"timestamp": -1}})
	var daRecords []daRecord.DataAgreementRecordForAuditList
	query := paginate.PaginateDBObjectsQueryUsingPipeline{
		Pipeline:   pipeline,
		Collection: daRecord.Collection(),
		Context:    context.Background(),
		Limit:      limit,
		Offset:     offset,
	}
	var resp fetchDataAgreementRecordsResp
	result, err := paginate.PaginateDBObjectsUsingPipeline(query, &daRecords)
	if err != nil {
		if errors.Is(err, paginate.EmptyDBError) {
			emptyDaRecords := make([]interface{}, 0)
			resp = fetchDataAgreementRecordsResp{
				DataAgreementRecords: emptyDaRecords,
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
