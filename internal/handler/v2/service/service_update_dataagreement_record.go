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

type updateconsentRecordReq struct {
	ConsentRecord consentRecord `json:"consentRecord"`
}

type consentRecord struct {
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

	// Request body
	var dataAgreementRecordReq updateDataAgreementRecordReq
	var updateConsentReq updateconsentRecordReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	var optIn bool
	// Unmarshal data agreement record req
	json.Unmarshal(b, &dataAgreementRecordReq)
	if dataAgreementRecordReq.OptIn {
		optIn = dataAgreementRecordReq.OptIn
	} else {
		// Unmarshal update consent record req
		json.Unmarshal(b, &updateConsentReq)
		if updateConsentReq.ConsentRecord.OptIn {
			optIn = updateConsentReq.ConsentRecord.OptIn
		} else {
			optIn = false
		}
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
	// Repository
	daRepo := dataagreement.DataAgreementRepository{}
	daRepo.Init(organisationId)

	_, err = daRepo.Get(toBeUpdatedDaRecord.DataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement: %v", toBeUpdatedDaRecord.DataAgreementId)
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	currentDataAgreementRecordRevision, err := revision.GetLatestByObjectIdAndSchemaName(dataAgreementRecordId, config.DataAgreementRecord)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch latest revision for data agreement record: %v", dataAgreementRecordId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	if toBeUpdatedDaRecord.OptIn == optIn {
		// response
		resp := updateDataAgreementRecordResp{
			DataAgreementRecord: toBeUpdatedDaRecord,
			Revision:            currentDataAgreementRecordRevision,
		}
		common.ReturnHTTPResponse(resp, w)
		return
	}
	toBeUpdatedDaRecord.OptIn = optIn

	currentDataAgreementRevision, err := revision.GetLatestByObjectIdAndSchemaName(toBeUpdatedDaRecord.DataAgreementId, config.DataAgreement)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch latest revision for data agreement: %v", toBeUpdatedDaRecord.DataAgreementId)
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
