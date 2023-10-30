package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/dataagreement"
	daRecord "github.com/bb-consent/api/src/v2/dataagreement_record"
	daRecordHistory "github.com/bb-consent/api/src/v2/dataagreement_record_history"
	"github.com/bb-consent/api/src/v2/revision"
	"github.com/bb-consent/api/src/v2/webhook"
	"github.com/bb-consent/api/src/webhooks"
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
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
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

	// Repository
	daRepo := dataagreement.DataAgreementRepository{}
	daRepo.Init(organisationId)

	da, err := daRepo.Get(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Trigger webhooks
	var consentedAttributes []string
	for _, pConsent := range da.DataAttributes {
		consentedAttributes = append(consentedAttributes, pConsent.Id.Hex())
	}

	go webhook.TriggerConsentWebhookEvent(individualId, dataAgreementId, savedDaRecord.Id.Hex(), organisationId, webhooks.EventTypes[30], strconv.FormatInt(time.Now().UTC().Unix(), 10), 0, consentedAttributes)
	// Add data agreement record history
	darH := daRecordHistory.DataAgreementRecordsHistory{}
	darH.DataAgreementId = dataAgreementId
	darH.OrganisationId = organisationId
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
