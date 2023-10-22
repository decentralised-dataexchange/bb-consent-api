package service

import (
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	daRecord "github.com/bb-consent/api/src/v2/dataagreement_record"
)

func ServiceDeleteIndividualDataAgreementRecords(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))
	individualId := common.Sanitize(r.Header.Get(config.IndividualHeaderKey))

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	err := darRepo.DeleteAllRecordsForIndividual(individualId)
	if err != nil {
		m := "Failed to delete data agreement records for individual"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

}
