package service

import (
	"log"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	daRecord "github.com/bb-consent/api/internal/dataagreement_record"
	"github.com/bb-consent/api/internal/paginate"
	"go.mongodb.org/mongo-driver/bson"
)

func dataAgreementRecordsToInterfaceSlice(dataAgreementRecords []daRecord.DataAgreementRecordWithTimestamp) []interface{} {
	interfaceSlice := make([]interface{}, len(dataAgreementRecords))
	for i, r := range dataAgreementRecords {
		interfaceSlice[i] = r
	}
	return interfaceSlice
}

type vFetchIndividualDataAgreementRecordsResp struct {
	DataAgreementRecords interface{}         `json:"consentRecords"`
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

	pipeline, err := daRecord.CreatePipelineForFilteringDataAgreementRecordsByIndividualId(organisationId, individualId)
	if err != nil {
		m := "Failed to create pipeline"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Return all data agreement records
	pipeline = append(pipeline, bson.M{"$sort": bson.M{"timestamp": -1}})

	dataAgreementRecords, err := daRecord.GetAllUsingPipeline(pipeline)
	if err != nil {
		m := "Failed to fetch data agreement records"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	darepo := dataagreement.DataAgreementRepository{}
	darepo.Init(organisationId)

	// filter consent records that are associated with non deleted data agreements
	var consentRecords []daRecord.DataAgreementRecordWithTimestamp
	for _, dataAgreementRecord := range dataAgreementRecords {
		_, err := darepo.Get(dataAgreementRecord.DataAgreementId)
		if err == nil {
			consentRecords = append(consentRecords, dataAgreementRecord)
		}
	}

	query := paginate.PaginateObjectsQuery{
		Limit:  limit,
		Offset: offset,
	}
	interfaceSlice := dataAgreementRecordsToInterfaceSlice(consentRecords)
	result := paginate.PaginateObjects(query, interfaceSlice)

	resp := vFetchIndividualDataAgreementRecordsResp{
		DataAgreementRecords: result.Items,
		Pagination:           result.Pagination,
	}
	common.ReturnHTTPResponse(resp, w)
}
