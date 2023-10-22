package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	daRecord "github.com/bb-consent/api/src/v2/dataagreement_record"
	"github.com/bb-consent/api/src/v2/revision"
	"github.com/bb-consent/api/src/v2/signature"
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
	toBeCreatedSignature, err := signature.CreateSignatureForObject("revision", daRecordRevision.Id.Hex(), false, daRecordRevision, false, signature.Signature{})
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
