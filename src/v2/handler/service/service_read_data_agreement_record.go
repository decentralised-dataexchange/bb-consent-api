package service

import (
	"encoding/json"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	daRecord "github.com/bb-consent/api/src/v2/dataagreement_record"
	"github.com/gorilla/mux"
)

type readDataAgreementRecordResp struct {
	DataAgreementRecord daRecord.DataAgreementRecord `json:"dataAgreementRecord"`
}

func ServiceReadDataAgreementRecord(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))
	individualId := common.Sanitize(r.Header.Get(config.IndividualHeaderKey))

	dataAgreementId := common.Sanitize(mux.Vars(r)[config.DataAgreementId])

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	daRecord, err := darRepo.GetByDataAgreementIdandIndividualId(dataAgreementId, individualId)
	if err != nil {
		m := "Failed to fetch data agreement record"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := readDataAgreementRecordResp{
		DataAgreementRecord: daRecord,
	}
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
