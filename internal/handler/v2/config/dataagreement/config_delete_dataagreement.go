package dataagreement

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	daRecord "github.com/bb-consent/api/internal/dataagreement_record"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/token"
	"github.com/gorilla/mux"
)

// ConfigDeleteDataAgreement
func ConfigDeleteDataAgreement(w http.ResponseWriter, r *http.Request) {
	// Current user
	orgAdminId := token.GetUserID(r)

	// Query params
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	dataAgreementId := mux.Vars(r)[config.DataAgreementId]
	dataAgreementId = common.Sanitize(dataAgreementId)

	// Repository
	daRepo := dataagreement.DataAgreementRepository{}
	daRepo.Init(organisationId)

	// Query data agreement by id
	toBeDeletedDA, err := daRepo.Get(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	var rev revision.Revision

	// If data agreement is published then:
	// a. Fetch latest revision
	if toBeDeletedDA.Active {
		rev, err = revision.GetLatestByObjectIdAndSchemaName(dataAgreementId, config.DataAgreement)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch revision: %v", dataAgreementId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

	} else {
		// Data agreement is draft
		// Create a revision on runtime
		rev, err = revision.CreateRevisionForDraftDataAgreement(toBeDeletedDA, orgAdminId)
		if err != nil {
			m := "Failed to create revision in run time"
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

	}

	// Mark data agreement as deleted
	toBeDeletedDA.IsDeleted = true

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	// Update deleted data agreement in db
	_, err = daRepo.Update(toBeDeletedDA)
	if err != nil {
		m := fmt.Sprintf("Failed to delete data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(rev)

	response, _ := json.Marshal(revisionForHTTPResponse)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
