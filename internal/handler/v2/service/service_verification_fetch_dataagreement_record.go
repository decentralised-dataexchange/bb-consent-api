package service

import (
	"encoding/json"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	daRecord "github.com/bb-consent/api/internal/dataagreement_record"
	"github.com/bb-consent/api/internal/revision"
	"github.com/gorilla/mux"
)

type vFetchDataAgreementRecordResp struct {
	DataAgreementRecord daRecord.DataAgreementRecord `json:"consentRecord"`
	Revision            revision.Revision            `json:"revision"`
}

func ServiceVerificationFetchDataAgreementRecord(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))

	consentRecordId := common.Sanitize(mux.Vars(r)[config.ConsentRecordId])

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	daRecord, err := darRepo.Get(consentRecordId)
	if err != nil {
		m := "Failed to fetch data agreement record"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	currentRevision, err := revision.GetLatestByObjectId(daRecord.Id)
	if err != nil {
		m := "Failed to fetch revision for data agreement record"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := vFetchDataAgreementRecordResp{
		DataAgreementRecord: daRecord,
		Revision:            currentRevision,
	}
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
