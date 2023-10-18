package dataagreement

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/dataagreement"
	"github.com/bb-consent/api/src/v2/dataattribute"
	"github.com/bb-consent/api/src/v2/revision"
	"github.com/gorilla/mux"
)

func deleteDataAgreementIdFromDataAttributes(dataAgreementId string, organisationId string) error {

	dataAttributeRepo := dataattribute.DataAttributeRepository{}
	dataAttributeRepo.Init(organisationId)

	err := dataAttributeRepo.RemoveDataAgreementIdFromDataAttributes(dataAgreementId)
	if err != nil {
		return err
	}

	return nil
}

// ConfigDeleteDataAgreement
func ConfigDeleteDataAgreement(w http.ResponseWriter, r *http.Request) {
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)
	dataAgreementId := mux.Vars(r)[config.DataAgreementId]
	dataAgreementId = common.Sanitize(dataAgreementId)

	// Repository
	daRepo := dataagreement.DataAgreementRepository{}
	daRepo.Init(organisationId)

	currentDataAgreement, err := daRepo.Get(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	currentRevision, err := revision.GetLatestByDataAgreementId(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch revisions: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	currentDataAgreement.IsDeleted = true

	// Remove the data agreement from data attribute
	err = deleteDataAgreementIdFromDataAttributes(dataAgreementId, organisationId)
	if err != nil {
		m := fmt.Sprintf("Failed to delete data agreement id from data attributes: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	_, err = daRepo.Update(currentDataAgreement)
	if err != nil {
		m := fmt.Sprintf("Failed to delete data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(currentRevision)

	response, _ := json.Marshal(revisionForHTTPResponse)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
