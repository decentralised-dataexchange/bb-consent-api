package audit

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

// AuditDataAgreementRecordRead
func AuditDataAgreementRecordRead(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))

	dataAgreementRecordId := common.Sanitize(mux.Vars(r)[config.DataAgreementRecordId])

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	// fetch data agreement record from db
	daRecord, err := darRepo.Get(dataAgreementRecordId)
	if err != nil {
		m := "Failed to fetch data agreement record"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// creating response
	resp := readDataAgreementRecordResp{
		DataAgreementRecord: daRecord,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}