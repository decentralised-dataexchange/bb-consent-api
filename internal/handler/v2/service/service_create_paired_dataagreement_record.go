package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	daRecord "github.com/bb-consent/api/internal/dataagreement_record"
	daRecordHistory "github.com/bb-consent/api/internal/dataagreement_record_history"
	"github.com/bb-consent/api/internal/individual"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/signature"
	"github.com/bb-consent/api/internal/webhook"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// createPairedDataAgreementRecord
func createPairedDataAgreementRecord(dataAgreementId string, rev revision.Revision, individualId string) daRecord.DataAgreementRecord {
	var newDaRecord daRecord.DataAgreementRecord

	newDaRecord.DataAgreementId = dataAgreementId
	newDaRecord.DataAgreementRevisionHash = rev.SerializedHash
	newDaRecord.DataAgreementRevisionId = rev.Id
	newDaRecord.IndividualId = individualId
	newDaRecord.OptIn = true
	newDaRecord.State = config.Unsigned

	return newDaRecord
}

type dataAgreementRecordReq struct {
	Id                        string `json:"id" bson:"_id,omitempty"`
	DataAgreementId           string `json:"dataAgreementId" valid:"required"`
	DataAgreementRevisionId   string `json:"dataAgreementRevisionId" valid:"required"`
	DataAgreementRevisionHash string `json:"dataAgreementRevisionHash"`
	IndividualId              string `json:"individualId" valid:"required"`
	OptIn                     bool   `json:"optIn"`
	State                     string `json:"state"`
	SignatureId               string `json:"signatureId"`
}

type createPairedDataAgreementRecordReq struct {
	DataAgreementRecord dataAgreementRecordReq `json:"consentRecord" valid:"required"`
	Signature           signature.Signature    `json:"signature"`
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

	// validating request payload
	valid, err := govalidator.ValidateStruct(dataAgreementRecordReq)
	if !valid {
		m := "Missing mandatory params for creating data agreement record"
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	individual, err := individualRepo.Get(dataAgreementRecordReq.DataAgreementRecord.IndividualId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch individual: %v", dataAgreementRecordReq.DataAgreementRecord.IndividualId)
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	// Check for existing data agreement record with same data agreement id and individual id
	count, err := darRepo.CountDataAgreementRecords(dataAgreementRecordReq.DataAgreementRecord.DataAgreementId, individual.Id)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement record for data agreement: %v", dataAgreementRecordReq.DataAgreementRecord.DataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	if count > 0 {
		m := fmt.Sprintf("Data agreement record for data agreement: %v and individual id : %s exists", dataAgreementRecordReq.DataAgreementRecord.DataAgreementId, individual.Id)
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	// Repository
	daRepo := dataagreement.DataAgreementRepository{}
	daRepo.Init(organisationId)

	dataAgreement, err := daRepo.Get(dataAgreementRecordReq.DataAgreementRecord.DataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement: %v", dataAgreementRecordReq.DataAgreementRecord.DataAgreementId)
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	// fetch revision based on id and schema name
	dataAgreementRevision, err := revision.GetByRevisionIdAndSchema(common.Sanitize(dataAgreementRecordReq.DataAgreementRecord.DataAgreementRevisionId), config.DataAgreement)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch revision: %v", dataAgreementRecordReq.DataAgreementRecord.DataAgreementRevisionId)
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	newDataAgreementRecord := createPairedDataAgreementRecord(dataAgreement.Id, dataAgreementRevision, individual.Id)

	dataAgreementRecord := newDataAgreementRecord
	dataAgreementRecord.OrganisationId = organisationId
	currentSignature := dataAgreementRecordReq.Signature
	dataAgreementRecord.Id = primitive.NewObjectID().Hex()
	currentSignature.Id = primitive.NewObjectID().Hex()
	dataAgreementRecord.SignatureId = currentSignature.Id

	newRecordRevision, err := revision.CreateRevisionForDataAgreementRecord(dataAgreementRecord, individualId)
	if err != nil {
		m := fmt.Sprintf("Failed to create new revision for dataAgreementRecord: %v", dataAgreementRecord.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	// create signature for data agreement record
	toBeCreatedSignature, err := signature.CreateSignatureForObject("revision", newRecordRevision.Id, false, newRecordRevision, true, currentSignature)
	if err != nil {
		m := fmt.Sprintf("Failed to create signature for data agreement record: %v", dataAgreementRecord.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

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
	darH.ConsentRecordId = savedDataAgreementRecord.Id
	darH.IndividualId = individual.Id
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
