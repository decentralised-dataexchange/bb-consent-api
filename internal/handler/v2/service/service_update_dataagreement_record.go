package service

import (
	"encoding/json"
	"errors"
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
	"github.com/bb-consent/api/internal/webhook"
	"github.com/gorilla/mux"
)

type updateDataAgreementRecordResp struct {
	DataAgreementRecord daRecord.DataAgreementRecord `json:"consentRecord"`
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

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	_, err := individualRepo.Get(individualId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch individual: %v", individualId)
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

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
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}
	// Repository
	daRepo := dataagreement.DataAgreementRepository{}
	daRepo.Init(organisationId)

	_, err = daRepo.Get(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
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

	currentDataAgreementRecordRevision, err := revision.GetLatestByObjectId(dataAgreementRecordId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch latest revision for data agreement record: %v", dataAgreementRecordId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	if toBeUpdatedDaRecord.OptIn == dataAgreementRecordReq.OptIn {
		// response
		resp := updateDataAgreementRecordResp{
			DataAgreementRecord: toBeUpdatedDaRecord,
			Revision:            currentDataAgreementRecordRevision,
		}
		common.ReturnHTTPResponse(resp, w)
		return
	}
	toBeUpdatedDaRecord.OptIn = dataAgreementRecordReq.OptIn

	currentDataAgreementRevision, err := revision.GetLatestByObjectId(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch latest revision for data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Create new revision
	newRevision, err := revision.UpdateRevisionForDataAgreementRecord(toBeUpdatedDaRecord, individualId, currentDataAgreementRevision)
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
	var eventType string
	if savedDaRecord.OptIn {
		eventType = webhook.EventTypes[30]

	} else {
		eventType = webhook.EventTypes[31]
	}

	go webhook.TriggerConsentWebhookEvent(savedDaRecord, organisationId, eventType)
	// Add data agreement record history
	darH := daRecordHistory.DataAgreementRecordsHistory{}
	darH.DataAgreementId = savedDaRecord.DataAgreementId
	darH.OrganisationId = organisationId
	darH.ConsentRecordId = savedDaRecord.Id
	darH.IndividualId = individualId
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
