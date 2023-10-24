package audit

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	daRecord "github.com/bb-consent/api/src/v2/dataagreement_record"
	"github.com/bb-consent/api/src/v2/paginate"
)

func dataAgreementRecordToInterfaceSlice(dataAgreementRecords []daRecord.DataAgreementRecordForAuditList) []interface{} {
	interfaceSlice := make([]interface{}, len(dataAgreementRecords))
	for i, r := range dataAgreementRecords {
		interfaceSlice[i] = r
	}
	return interfaceSlice
}

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

	var isNotDataAgreementRecordId bool
	var isNotIndividualId bool
	var isNotDataAgreementId bool
	var isNotLawfulBasis bool
	dataAgreementRecordId, err := daRecord.ParseQueryParams(r, config.DataAgreementRecordId, daRecord.DataAgreementRecordIdIsMissingError)
	if err != nil && errors.Is(err, daRecord.DataAgreementRecordIdIsMissingError) {
		isNotDataAgreementRecordId = true
	}
	dataAgreementId, err := daRecord.ParseQueryParams(r, config.DataAgreementId, daRecord.DataAgreementIdIsMissingError)
	if err != nil && errors.Is(err, daRecord.DataAgreementIdIsMissingError) {
		isNotDataAgreementId = true
	}
	individualId, err := daRecord.ParseQueryParams(r, config.IndividualId, daRecord.IndividualIdIsMissingError)
	if err != nil && errors.Is(err, daRecord.IndividualIdIsMissingError) {
		isNotIndividualId = true
	}

	lawfulBasis, err := daRecord.ParseQueryParams(r, config.LawfulBasis, daRecord.LawfulBasisIsMissingError)
	if err != nil && errors.Is(err, daRecord.LawfulBasisIsMissingError) {
		isNotLawfulBasis = true
	}

	var daRecords []daRecord.DataAgreementRecordForAuditList
	if isNotLawfulBasis {
		if isNotDataAgreementRecordId {
			if isNotIndividualId {
				if isNotDataAgreementId {
					m := "Query params missing"
					common.HandleErrorV2(w, http.StatusInternalServerError, m, errors.New("invalid query params"))
					return

				} else {
					// fetch by data agreement id
					daRecords, err = daRecord.ListByDataAgreementIdIncludingDataAgreement(dataAgreementId, organisationId)
					if err != nil {
						m := "Failed to fetch data agreement record"
						common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
						return
					}
				}

			} else {
				// fetch by individual id
				daRecords, err = daRecord.ListByIndividualIdIncludingDataAgreement(individualId, organisationId)
				if err != nil {
					m := "Failed to fetch data agreement record"
					common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
					return
				}
			}

		} else {
			// fetch by data agreement record id
			daRecords, err = daRecord.ListByIdIncludingDataAgreement(dataAgreementRecordId, organisationId)
			if err != nil {
				m := "Failed to fetch data agreement record"
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}

		}

	} else {
		if isNotDataAgreementRecordId {
			if isNotIndividualId {
				// fetch by data agreement id and lawful basis
				daRecords, err = daRecord.ListByDataAgreementIdAndLawfulBasis(dataAgreementId, organisationId, lawfulBasis)
				if err != nil {
					m := "Failed to fetch data agreement record"
					common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
					return
				}
			} else {
				// fetch by individual id and lawfulusage
				daRecords, err = daRecord.ListByIndividualIdAndLawfulBasis(individualId, organisationId, lawfulBasis)
				if err != nil {
					m := "Failed to fetch data agreement record"
					common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
					return
				}
			}

		} else {
			// fetch by data agreement record id and lawful basis
			daRecords, err = daRecord.ListByIdAndLawfulBasis(dataAgreementRecordId, organisationId, lawfulBasis)
			if err != nil {
				m := "Failed to fetch data agreement record"
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}

		}

	}
	query := paginate.PaginateObjectsQuery{
		Limit:  limit,
		Offset: offset,
	}
	interfaceSlice := dataAgreementRecordToInterfaceSlice(daRecords)
	result := paginate.PaginateObjects(query, interfaceSlice)

	var resp = fetchDataAgreementRecordsResp{
		DataAgreementRecords: result.Items,
		Pagination:           result.Pagination,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
