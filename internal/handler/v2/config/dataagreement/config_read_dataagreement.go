package dataagreement

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

	// Query data agreement by id
	da, err := daRepo.Get(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	var rev revision.Revision

	// If `revisionId` query param is provided
	// then query revision by id
	if strings.TrimSpace(revisionId) != "" {

		rev, err = revision.GetByRevisionId(revisionId)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch revision: %v", dataAgreementId)
			common.HandleErrorV2(w, http.StatusBadRequest, m, err)
			return
		}

	} else {
		// If `revisionId` query param is not provided
		// If data agreement is published then:
		// a. Fetch latest revision
		if da.Active {
			rev, err = revision.GetLatestByObjectIdAndSchemaName(dataAgreementId, config.DataAgreement)
			if err != nil {
				m := fmt.Sprintf("Failed to fetch revision: %v", dataAgreementId)
				common.HandleErrorV2(w, http.StatusBadRequest, m, err)
				return
			}

		} else {
			// Data agreement is draft
			// Create a revision on runtime
			rev, err = revision.CreateRevisionForDraftDataAgreement(da, orgAdminId)
			if err != nil {
				m := "Failed to create revision in run time"
				common.HandleErrorV2(w, http.StatusBadRequest, m, err)
				return
			}

		}
	}

	// Constructing the response
	var resp getDataAgreementResp
	resp.DataAgreement = da

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(rev)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
