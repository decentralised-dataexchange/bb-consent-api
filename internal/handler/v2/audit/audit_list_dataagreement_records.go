package audit

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	daRecord "github.com/bb-consent/api/internal/dataagreement_record"
	"github.com/bb-consent/api/internal/paginate"
	"github.com/bb-consent/api/internal/revision"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type dataAgreementForListDataAgreementRecord struct {
	Purpose     string `json:"purpose"`
	LawfulBasis string `json:"lawfulBasis"`
	Version     string `json:"version"`
}

type listDataAgreementRecord struct {
	Id                        primitive.ObjectID                      `json:"id" bson:"_id,omitempty"`
	DataAgreementId           string                                  `json:"dataAgreementId"`
	DataAgreementRevisionId   string                                  `json:"dataAgreementRevisionId"`
	DataAgreementRevisionHash string                                  `json:"dataAgreementRevisionHash"`
	IndividualId              string                                  `json:"individualId"`
	OptIn                     bool                                    `json:"optIn"`
	State                     string                                  `json:"state" valid:"required"`
	SignatureId               string                                  `json:"signatureId"`
	Timestamp                 string                                  `json:"timestamp"`
	DataAgreement             dataAgreementForListDataAgreementRecord `json:"dataAgreement"`
}

func dataAgreementRecordsToInterfaceSlice(dataAgreementRecords []listDataAgreementRecord) []interface{} {
	interfaceSlice := make([]interface{}, len(dataAgreementRecords))
	for i, r := range dataAgreementRecords {
		interfaceSlice[i] = r
	}
	return interfaceSlice
}

// recreateDataAgreementRecordsFromRevisions
func recreateDataAgreementRecordsFromRevisions(dataAgreementRecords []daRecord.DataAgreementRecordForAuditList, lawfulbasis string) ([]listDataAgreementRecord, error) {
	var consentRecords []listDataAgreementRecord

	for _, dARecord := range dataAgreementRecords {
		for _, dARevision := range dARecord.Revisions {
			var consentRecord listDataAgreementRecord
			// recreate consent record from revisions objectdata
			tempDARecord, err := revision.RecreateConsentRecordFromObjectData(dARevision.ObjectData)
			if err != nil {
				return consentRecords, err
			}
			// populate the data agreement revision values
			consentRecord.Id = tempDARecord.Id
			consentRecord.DataAgreementId = tempDARecord.DataAgreementId
			consentRecord.DataAgreementRevisionHash = tempDARecord.DataAgreementRevisionHash
			consentRecord.DataAgreementRevisionId = tempDARecord.DataAgreementRevisionId
			consentRecord.IndividualId = tempDARecord.IndividualId
			consentRecord.OptIn = tempDARecord.OptIn
			consentRecord.State = tempDARecord.State
			consentRecord.SignatureId = tempDARecord.SignatureId
			consentRecord.Timestamp = dARevision.Timestamp
			// fetch corresponding data agreement revision
			dataAgreementRevision, err := revision.GetByRevisionId(tempDARecord.DataAgreementRevisionId)
			if err != nil {
				return consentRecords, err
			}
			// recreate data agreement from revision
			dataAgreement, err := recreateDataAgreementFromObjectData(dataAgreementRevision.ObjectData)
			if err != nil {
				return consentRecords, err
			}
			// populate data agreement values obtained after recreating data agreement
			consentRecord.DataAgreement.LawfulBasis = dataAgreement.LawfulBasis
			consentRecord.DataAgreement.Purpose = dataAgreement.Purpose
			consentRecord.DataAgreement.Version = dataAgreement.Version
			// filter by lawful basis
			if len(lawfulbasis) > 0 {
				if consentRecord.DataAgreement.LawfulBasis == lawfulbasis {
					consentRecords = append(consentRecords, consentRecord)
				}
			} else {
				consentRecords = append(consentRecords, consentRecord)
			}

		}
	}
	return consentRecords, nil
}

func recreateDataAgreementFromObjectData(objectData string) (dataAgreementForListDataAgreementRecord, error) {

	// Deserialise data agreement
	var da dataAgreementForListDataAgreementRecord
	err := json.Unmarshal([]byte(objectData), &da)
	if err != nil {
		return da, err
	}

	return da, nil
}

type fetchDataAgreementRecordsResp struct {
	DataAgreementRecords interface{}         `json:"consentRecords"`
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

	dataAgreementRecords, err := daRecord.DataAgreementRecordsWithRevisionsFilteredById(organisationId, id)
	if err != nil {
		m := "Failed to fetch all data agreement records"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	consentRecords, err := recreateDataAgreementRecordsFromRevisions(dataAgreementRecords, lawfulBasis)
	if err != nil {
		m := "Failed to recreate data agreement records from revisions"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	query := paginate.PaginateObjectsQuery{
		Limit:  limit,
		Offset: offset,
	}
	interfaceSlice := dataAgreementRecordsToInterfaceSlice(consentRecords)
	result := paginate.PaginateObjects(query, interfaceSlice)

	resp := fetchDataAgreementRecordsResp{
		DataAgreementRecords: result.Items,
		Pagination:           result.Pagination,
	}
	common.ReturnHTTPResponse(resp, w)

}
