package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/dataagreement"
	"github.com/bb-consent/api/src/v2/revision"
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

	// Parse the URL query parameters
	revisionId := r.URL.Query().Get("revisionId")
	revisionId = common.Sanitize(revisionId)

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

	if revisionId != "" {

		revisionResp, err = revision.GetByRevisionId(revisionId)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch revision: %v", dataAgreementId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

	} else {
		revisionResp, err = revision.GetLatestByDataAgreementId(dataAgreementId)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch revision: %v", dataAgreementId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
	}

	// Constructing the response
	var resp getDataAgreementResp
	resp.DataAgreement = da

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(revisionResp)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
