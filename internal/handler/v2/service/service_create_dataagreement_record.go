package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	daRecord "github.com/bb-consent/api/internal/dataagreement_record"
	daRecordHistory "github.com/bb-consent/api/internal/dataagreement_record_history"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/webhook"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// createDataAgreementRecord
func createDataAgreementRecord(dataAgreementId string, rev revision.Revision, individualId string) daRecord.DataAgreementRecord {
	var newDaRecord daRecord.DataAgreementRecord

	newDaRecord.Id = primitive.NewObjectID()
	newDaRecord.DataAgreementId = dataAgreementId
	newDaRecord.DataAgreementRevisionHash = rev.SerializedHash
	newDaRecord.DataAgreementRevisionId = rev.Id.Hex()
	newDaRecord.IndividualId = individualId
	newDaRecord.OptIn = true
	newDaRecord.State = config.Unsigned

	return newDaRecord
}

type createDataAgreementRecordResp struct {
	DataAgreementRecord daRecord.DataAgreementRecord `json:"consentRecord"`
	Revision            revision.Revision            `json:"revision"`
}

// ServiceCreateDataAgreementRecord
func ServiceCreateDataAgreementRecord(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))
	individualId := common.Sanitize(r.Header.Get(config.IndividualHeaderKey))

	dataAgreementId := common.Sanitize(mux.Vars(r)[config.DataAgreementId])

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	// Check for existing data agreement record with same data agreement id and individual id
	count, err := darRepo.CountDataAgreementRecords(dataAgreementId, individualId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement record for data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	if count > 0 {
		m := fmt.Sprintf("Data agreement record for data agreement: %v and individual id : %s exists", dataAgreementId, individualId)
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	revisionId, err := daRecord.ParseQueryParams(r, config.RevisionId, daRecord.RevisionIdIsMissingError)
	revisionId = common.Sanitize(revisionId)
	var rev revision.Revision

	// If revision id is missing, fetch latest revision
	if err != nil && errors.Is(err, daRecord.RevisionIdIsMissingError) {
		rev, err = revision.GetLatestByObjectId(dataAgreementId)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch revision for data agreement: %v", dataAgreementId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
	} else {
		// fetch revision based on id and schema name
		rev, err = revision.GetByRevisionIdAndSchema(revisionId, config.DataAgreement)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch revision: %v", revisionId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
	}

	// create new data agreement record
	newDaRecord := createDataAgreementRecord(dataAgreementId, rev, individualId)
	newDaRecord.OrganisationId = organisationId
	newDaRecord.IsDeleted = false

	// Create new revision
	newRevision, err := revision.CreateRevisionForDataAgreementRecord(newDaRecord, individualId)
	if err != nil {
		m := "Failed to create revision for new data agreement record"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	savedDaRecord, err := darRepo.Add(newDaRecord)
	if err != nil {
		m := "Failed to create new data agreement record"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Save the revision to db
	savedRevision, err := revision.Add(newRevision)
	if err != nil {
		m := fmt.Sprintf("Failed to create new revision: %v", newRevision.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Trigger webhooks
	go webhook.TriggerConsentWebhookEvent(savedDaRecord, organisationId, webhook.EventTypes[30])

	// Add data agreement record history
	darH := daRecordHistory.DataAgreementRecordsHistory{}
	darH.DataAgreementId = dataAgreementId
	darH.OrganisationId = organisationId
	darH.ConsentRecordId = savedDaRecord.Id.Hex()
	darH.IndividualId = individualId
	err = daRecordHistory.DataAgreementRecordHistoryAdd(darH, savedDaRecord.OptIn)
	if err != nil {
		m := "Failed to add data agreement record history"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// response
	resp := createDataAgreementRecordResp{
		DataAgreementRecord: savedDaRecord,
		Revision:            savedRevision,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
