package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	daRecord "github.com/bb-consent/api/src/v2/dataagreement_record"
	daRecordHistory "github.com/bb-consent/api/src/v2/dataagreement_record_history"
	"github.com/bb-consent/api/src/v2/revision"
	"github.com/bb-consent/api/src/v2/signature"
	"github.com/bb-consent/api/src/v2/webhook"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type createPairedDataAgreementRecordReq struct {
	DataAgreementRecord daRecord.DataAgreementRecord `json:"dataAgreementRecord" valid:"required"`
	Signature           signature.Signature          `json:"signature" valid:"required"`
}

type createPairedDataAgreementRecordResp struct {
	DataAgreementRecord daRecord.DataAgreementRecord `json:"dataAgreementRecord"`
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

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

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

	// Trigger webhooks
	var consentedAttributes []string
	for _, pConsent := range savedDataAgreementRecord.DataAttributes {
		consentedAttributes = append(consentedAttributes, pConsent.DataAttributeId)
	}
	var eventType string
	if savedDataAgreementRecord.OptIn {
		eventType = webhook.EventTypes[30]

	} else {
		eventType = webhook.EventTypes[30]
	}

	go webhook.TriggerConsentWebhookEvent(individualId, savedDataAgreementRecord.DataAgreementId, savedDataAgreementRecord.Id.Hex(), organisationId, eventType, strconv.FormatInt(time.Now().UTC().Unix(), 10), 0, consentedAttributes)

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
