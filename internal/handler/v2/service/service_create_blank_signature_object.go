package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	daRecord "github.com/bb-consent/api/internal/dataagreement_record"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/signature"
	"github.com/gorilla/mux"
)

type createBlankSignatureResp struct {
	Signature signature.Signature `json:"signature"`
}

func ServiceCreateBlankSignature(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))

	dataAgreementRecordId := common.Sanitize(mux.Vars(r)[config.DataAgreementRecordId])

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	// Get latest revision for data agreement record
	daRecordRevision, err := revision.GetLatestByObjectId(dataAgreementRecordId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch revision for data agreement record: %v", dataAgreementRecordId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	// create signature for data agreement record
	toBeCreatedSignature, err := signature.CreateSignatureForObject("revision", daRecordRevision.Id, false, daRecordRevision, false, signature.Signature{})
	if err != nil {
		m := fmt.Sprintf("Failed to create signature for data agreement record: %v", dataAgreementRecordId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Creating response
	resp := createBlankSignatureResp{
		Signature: toBeCreatedSignature,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
