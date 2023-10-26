package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	daRecord "github.com/bb-consent/api/src/v2/dataagreement_record"
	daRecordHistory "github.com/bb-consent/api/src/v2/dataagreement_record_history"
	"github.com/bb-consent/api/src/v2/revision"
	"github.com/bb-consent/api/src/v2/webhook"
	"github.com/gorilla/mux"
)

// updateOptInOfDataAttributes
func updateOptInOfDataAttributes(daR daRecord.DataAgreementRecord) daRecord.DataAgreementRecord {
	var dataAttributes []daRecord.DataAttributeForDataAgreementRecord
	for _, da := range daR.DataAttributes {
		da.OptIn = daR.OptIn
		dataAttributes = append(dataAttributes, da)
	}
	daR.DataAttributes = dataAttributes
	return daR
}

type updateDataAgreementRecordResp struct {
	DataAgreementRecord daRecord.DataAgreementRecord `json:"dataAgreementRecord"`
	Revision            revision.Revision            `json:"revision"`
}
type updateDataAgreementRecordReq struct {
	OptIn bool `json:"optIn"`
}

func ServiceUpdateDataAgreementRecord(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))
	individualId := common.Sanitize(r.Header.Get(config.IndividualHeaderKey))

	dataAgreementRecordId := common.Sanitize(mux.Vars(r)[config.DataAgreementRecordId])

	// Parse query params
	dataAgreementId, err := daRecord.ParseQueryParams(r, config.DataAgreementId, daRecord.DataAgreementIdIsMissingError)
	dataAgreementId = common.Sanitize(dataAgreementId)
	if err != nil && errors.Is(err, daRecord.DataAgreementIdIsMissingError) {
		m := "Query param dataAgreementId is required"
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	// Request body
	var dataAgreementRecordReq updateDataAgreementRecordReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &dataAgreementRecordReq)

	// validating request payload
	valid, err := govalidator.ValidateStruct(dataAgreementRecordReq)
	if !valid {
		m := fmt.Sprintf("Failed to validate request body: %v", dataAgreementRecordId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	toBeUpdatedDaRecord, err := darRepo.Get(dataAgreementRecordId)
	if err != nil {
		m := "Failed to fetch data agreement record"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	if toBeUpdatedDaRecord.OptIn == dataAgreementRecordReq.OptIn {
		m := "Data agreement record opt in is same as provided value"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	toBeUpdatedDaRecord.OptIn = dataAgreementRecordReq.OptIn
	toBeUpdatedDaRecord = updateOptInOfDataAttributes(toBeUpdatedDaRecord)

	currentDataAgreementRecordRevision, err := revision.GetLatestByObjectId(dataAgreementRecordId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch latest revision for data agreement record: %v", dataAgreementRecordId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Create new revision
	newRevision, err := revision.UpdateRevisionForDataAgreementRecord(toBeUpdatedDaRecord, &currentDataAgreementRecordRevision, individualId)
	if err != nil {
		m := "Failed to create revision for new data agreement record"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	savedDaRecord, err := darRepo.Update(toBeUpdatedDaRecord)
	if err != nil {
		m := "Failed to update new data agreement record"
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
	var consentedAttributes []string
	for _, pConsent := range savedDaRecord.DataAttributes {
		consentedAttributes = append(consentedAttributes, pConsent.DataAttributeId)
	}
	var eventType string
	if savedDaRecord.OptIn {
		eventType = webhook.EventTypes[30]

	} else {
		eventType = webhook.EventTypes[30]
	}

	go webhook.TriggerConsentWebhookEvent(individualId, dataAgreementId, dataAgreementRecordId, organisationId, eventType, strconv.FormatInt(time.Now().UTC().Unix(), 10), 0, consentedAttributes)
	// Add data agreement record history
	darH := daRecordHistory.DataAgreementRecordsHistory{}
	darH.DataAgreementId = savedDaRecord.DataAgreementId
	darH.OrganisationId = organisationId
	err = daRecordHistory.DataAgreementRecordHistoryAdd(darH, savedDaRecord.OptIn)
	if err != nil {
		m := "Failed to add data agreement record history"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// response
	resp := updateDataAgreementRecordResp{
		DataAgreementRecord: savedDaRecord,
		Revision:            savedRevision,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
