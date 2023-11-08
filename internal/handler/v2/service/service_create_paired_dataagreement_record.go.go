package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	daRecord "github.com/bb-consent/api/internal/dataagreement_record"
	daRecordHistory "github.com/bb-consent/api/internal/dataagreement_record_history"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/signature"
	"github.com/bb-consent/api/internal/webhook"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type createPairedDataAgreementRecordReq struct {
	DataAgreementRecord daRecord.DataAgreementRecord `json:"consentRecord" valid:"required"`
	Signature           signature.Signature          `json:"signature" valid:"required"`
}

type createPairedDataAgreementRecordResp struct {
	DataAgreementRecord daRecord.DataAgreementRecord `json:"consentRecord"`
	Revision            revision.Revision            `json:"revision"`
	Signature           signature.Signature          `json:"signature"`
}

func ServiceCreatePairedDataAgreementRecord(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))
	individualId := common.Sanitize(r.Header.Get(config.IndividualHeaderKey))

	// Request body
	var dataAgreementRecordReq createPairedDataAgreementRecordReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &dataAgreementRecordReq)

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	// Check for existing data agreement record with same data agreement id and individual id
	count, err := darRepo.CountDataAgreementRecords(dataAgreementRecordReq.DataAgreementRecord.DataAgreementId, individualId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement record for data agreement: %v", dataAgreementRecordReq.DataAgreementRecord.DataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	if count > 0 {
		m := fmt.Sprintf("Data agreement record for data agreement: %v and individual id : %s exists", dataAgreementRecordReq.DataAgreementRecord.DataAgreementId, individualId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	dataAgreementRecord := dataAgreementRecordReq.DataAgreementRecord
	currentSignature := dataAgreementRecordReq.Signature

	dataAgreementRecord.Id = primitive.NewObjectID()

	newRecordRevision, err := revision.CreateRevisionForDataAgreementRecord(dataAgreementRecord, individualId)
	if err != nil {
		m := fmt.Sprintf("Failed to create new revision for dataAgreementRecord: %v", dataAgreementRecord.Id.Hex())
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	// create signature for data agreement record
	toBeCreatedSignature, err := signature.CreateSignatureForObject("revision", newRecordRevision.Id.Hex(), false, newRecordRevision, true, currentSignature)
	if err != nil {
		m := fmt.Sprintf("Failed to create signature for data agreement record: %v", dataAgreementRecord.Id.Hex())
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	toBeCreatedSignature.Id = primitive.NewObjectID()

	dataAgreementRecord.SignatureId = toBeCreatedSignature.Id.Hex()

	savedDataAgreementRecord, err := darRepo.Add(dataAgreementRecord)
	if err != nil {
		m := fmt.Sprintf("Failed to update paired data agreement record: %v", savedDataAgreementRecord.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	savedRevision, err := revision.Add(newRecordRevision)
	if err != nil {
		m := fmt.Sprintf("Failed to add revision for data agreement record: %v", savedDataAgreementRecord.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	savedSignature, err := signature.Add(toBeCreatedSignature)
	if err != nil {
		m := fmt.Sprintf("Failed to add signature for data agreement record: %v", savedDataAgreementRecord.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Repository
	daRepo := dataagreement.DataAgreementRepository{}
	daRepo.Init(organisationId)

	_, err = daRepo.Get(savedDataAgreementRecord.DataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement: %v", savedDataAgreementRecord.DataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Trigger webhooks
	var eventType string
	if savedDataAgreementRecord.OptIn {
		eventType = webhook.EventTypes[30]

	} else {
		eventType = webhook.EventTypes[31]
	}

	go webhook.TriggerConsentWebhookEvent(savedDataAgreementRecord, organisationId, eventType)

	// Add data agreement record history
	darH := daRecordHistory.DataAgreementRecordsHistory{}
	darH.DataAgreementId = dataAgreementRecord.DataAgreementId
	darH.OrganisationId = organisationId
	err = daRecordHistory.DataAgreementRecordHistoryAdd(darH, savedDataAgreementRecord.OptIn)
	if err != nil {
		m := "Failed to add data agreement record history"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	// response
	resp := createPairedDataAgreementRecordResp{
		DataAgreementRecord: savedDataAgreementRecord,
		Revision:            savedRevision,
		Signature:           savedSignature,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
