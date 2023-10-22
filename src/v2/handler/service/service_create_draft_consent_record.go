package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	daRecord "github.com/bb-consent/api/src/v2/dataagreement_record"
	"github.com/bb-consent/api/src/v2/dataattribute"
	"github.com/bb-consent/api/src/v2/revision"
	"github.com/bb-consent/api/src/v2/signature"
)

// getDataAttributesWithRevision
func getDataAttributesWithRevision(dataAttributes []dataattribute.DataAttribute) ([]daRecord.DataAttributeForDataAgreementRecord, error) {
	var dataAttributesWithRevision []daRecord.DataAttributeForDataAgreementRecord

	for _, da := range dataAttributes {
		var dataAttributeWithRevision daRecord.DataAttributeForDataAgreementRecord

		dataAttributeWithRevision.DataAttributeId = da.Id.Hex()
		revisionForDataAttibute, err := revision.GetLatestByDataAttributeId(da.Id.Hex())
		if err != nil {
			return dataAttributesWithRevision, err
		}
		dataAttributeWithRevision.DataAttributeRevisionId = revisionForDataAttibute.Id.Hex()
		dataAttributeWithRevision.DataAttributeRevisionHash = revisionForDataAttibute.SerializedHash
		dataAttributeWithRevision.OptIn = true

		dataAttributesWithRevision = append(dataAttributesWithRevision, dataAttributeWithRevision)
	}
	return dataAttributesWithRevision, nil
}

// createDraftDataAgreementRecord
func createDraftDataAgreementRecord(dataAgreementId string, rev revision.Revision, individualId string, dataAttributesWithRevision []daRecord.DataAttributeForDataAgreementRecord) daRecord.DataAgreementRecord {
	var newDaRecord daRecord.DataAgreementRecord

	newDaRecord.DataAgreementId = dataAgreementId
	newDaRecord.DataAgreementRevisionHash = rev.SerializedHash
	newDaRecord.DataAgreementRevisionId = rev.Id.Hex()
	newDaRecord.DataAttributes = dataAttributesWithRevision
	newDaRecord.IndividualId = individualId
	newDaRecord.OptIn = true
	newDaRecord.State = config.Unsigned

	return newDaRecord
}

type draftDataAgreementRecordResp struct {
	DataAgreementRecord daRecord.DataAgreementRecord `json:"dataAgreementRecord"`
	Signature           signature.Signature          `json:"signature"`
}

func ServiceCreateDraftConsentRecord(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))
	individualId := common.Sanitize(r.Header.Get(config.IndividualHeaderKey))

	// Parse query params
	dataAgreementId, err := daRecord.ParseQueryParams(r, config.DataAgreementId, daRecord.DataAgreementIdIsMissingError)
	dataAgreementId = common.Sanitize(dataAgreementId)
	if err != nil && errors.Is(err, daRecord.DataAgreementIdIsMissingError) {
		m := "Query param dataAgreementId is required"
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	revisionId, err := daRecord.ParseQueryParams(r, config.RevisionId, daRecord.RevisionIdIsMissingError)
	revisionId = common.Sanitize(revisionId)
	var rev revision.Revision

	// If revision id is missing, fetch latest revision
	if err != nil && errors.Is(err, daRecord.RevisionIdIsMissingError) {
		rev, err = revision.GetLatestByDataAgreementId(dataAgreementId)
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

	// Repository
	dataAttributeRepo := dataattribute.DataAttributeRepository{}
	dataAttributeRepo.Init(organisationId)

	// Fetch data attributes of data agreement from db
	dataAttributes, err := dataAttributeRepo.GetDataAttributesByDataAgreementId(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data attributes for data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Fecth revisions for data attributes
	dataAttributesWithRevision, err := getDataAttributesWithRevision(dataAttributes)
	if err != nil {
		m := "Failed to fetch revisions for data attributes"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// create new draft data agreement record
	newDaRecord := createDraftDataAgreementRecord(dataAgreementId, rev, individualId, dataAttributesWithRevision)

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
