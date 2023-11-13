package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	daRecord "github.com/bb-consent/api/internal/dataagreement_record"
	"github.com/bb-consent/api/internal/individual"
	"github.com/gorilla/mux"
)

type readDataAgreementRecordResp struct {
	DataAgreementRecord interface{} `json:"consentRecord"`
}

func ServiceReadDataAgreementRecord(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))
	individualId := common.Sanitize(r.Header.Get(config.IndividualHeaderKey))

	dataAgreementId := common.Sanitize(mux.Vars(r)[config.DataAgreementId])

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	// fetch the individual
	_, err := individualRepo.Get(individualId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch individual: %v", individualId)
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	var consentRecord interface{}
	consentRecord, err = darRepo.GetByDataAgreementIdandIndividualId(dataAgreementId, individualId)
	if err != nil {
		consentRecord = nil
	}

	resp := readDataAgreementRecordResp{
		DataAgreementRecord: consentRecord,
	}
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
