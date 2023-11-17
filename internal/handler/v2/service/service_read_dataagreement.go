package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/revision"
	"github.com/gorilla/mux"
)

type getDataAgreementResp struct {
	DataAgreement dataagreement.DataAgreement `json:"dataAgreement"`
	Revision      interface{}                 `json:"revision"`
}

// ServiceReadDataAgreement
func ServiceReadDataAgreement(w http.ResponseWriter, r *http.Request) {
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	dataAgreementId := mux.Vars(r)[config.DataAgreementId]
	dataAgreementId = common.Sanitize(dataAgreementId)

	// Repository
	daRepo := dataagreement.DataAgreementRepository{}
	daRepo.Init(organisationId)

	da, err := daRepo.Get(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	var revisionResp revision.Revision

	revisionResp, err = revision.GetLatestByDataAgreementId(da.Id.Hex())
	if err != nil {
		m := fmt.Sprintf("Failed to fetch revision: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	dataAgreement, err := revision.RecreateDataAgreementFromRevision(revisionResp)
	if err != nil {
		m := fmt.Sprintf("Failed to recreate data agreement from revision: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Constructing the response
	var resp getDataAgreementResp
	resp.DataAgreement = dataAgreement

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(revisionResp)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
