package dataagreement

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/token"
	"github.com/gorilla/mux"
)

type getDataAgreementResp struct {
	DataAgreement dataagreement.DataAgreement `json:"dataAgreement"`
	Revision      interface{}                 `json:"revision"`
}

// ConfigReadDataAgreement
func ConfigReadDataAgreement(w http.ResponseWriter, r *http.Request) {
	// Current user
	orgAdminId := token.GetUserID(r)

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
		if da.Version == "" {
			revisionResp, err = revision.CreateRevisionForDraftDataAgreement(da, orgAdminId)
			if err != nil {
				m := fmt.Sprintf("Failed to create revision for draft data agreement: %v", dataAgreementId)
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
