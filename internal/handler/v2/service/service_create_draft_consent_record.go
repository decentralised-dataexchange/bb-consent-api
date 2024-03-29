package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	daRecord "github.com/bb-consent/api/internal/dataagreement_record"
	"github.com/bb-consent/api/internal/individual"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/signature"
)

// createDraftDataAgreementRecord
func createDraftDataAgreementRecord(dataAgreementId string, rev revision.Revision, individualId string) daRecord.DataAgreementRecord {
	var newDaRecord daRecord.DataAgreementRecord

	newDaRecord.DataAgreementId = dataAgreementId
	newDaRecord.DataAgreementRevisionHash = rev.SerializedHash
	newDaRecord.DataAgreementRevisionId = rev.Id
	newDaRecord.IndividualId = individualId
	newDaRecord.OptIn = true
	newDaRecord.State = config.Unsigned

	return newDaRecord
}

type draftDataAgreementRecordResp struct {
	DataAgreementRecord daRecord.DataAgreementRecord `json:"consentRecord"`
	Signature           signature.Signature          `json:"signature"`
}

func ServiceCreateDraftConsentRecord(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))

	// Parse query params
	dataAgreementId, err := daRecord.ParseQueryParams(r, config.DataAgreementId, daRecord.DataAgreementIdIsMissingError)
	dataAgreementId = common.Sanitize(dataAgreementId)
	if err != nil && errors.Is(err, daRecord.DataAgreementIdIsMissingError) {
		m := "Query param dataAgreementId is required"
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}
	individualId, err := daRecord.ParseQueryParams(r, config.IndividualId, daRecord.IndividualIdIsMissingError)
	individualId = common.Sanitize(individualId)
	if err != nil && errors.Is(err, daRecord.IndividualIdIsMissingError) {
		m := "Query param individualId is required"
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}
	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	_, err = individualRepo.Get(individualId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch individual: %v", individualId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	revisionId, err := daRecord.ParseQueryParams(r, config.RevisionId, daRecord.RevisionIdIsMissingError)
	revisionId = common.Sanitize(revisionId)
	var rev revision.Revision

	// If revision id is missing, fetch latest revision
	if err != nil && errors.Is(err, daRecord.RevisionIdIsMissingError) {
		rev, err = revision.GetLatestByObjectIdAndSchemaName(dataAgreementId, config.DataAgreement)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch revision: %v", revisionId)
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

	// create new draft data agreement record
	newDaRecord := createDraftDataAgreementRecord(dataAgreementId, rev, individualId)

	// response
	resp := draftDataAgreementRecordResp{
		DataAgreementRecord: newDaRecord,
		Signature:           signature.Signature{},
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
