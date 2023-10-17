package dataattribute

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/dataattribute"
	"github.com/bb-consent/api/src/revision"
	"github.com/gorilla/mux"
)

type getDataAttributeResp struct {
	DataAttribute dataattribute.DataAttribute `json:"dataAttribute"`
	Revision      interface{}                 `json:"revision"`
}

// ConfigReadDataAttribute
func ConfigReadDataAttribute(w http.ResponseWriter, r *http.Request) {

	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	dataAttributeId := mux.Vars(r)[config.DataAttributeId]
	dataAttributeId = common.Sanitize(dataAttributeId)

	// Parse the URL query parameters
	revisionId := r.URL.Query().Get("revisionId")
	revisionId = common.Sanitize(revisionId)

	// Repository
	dataAttributeRepo := dataattribute.DataAttributeRepository{}
	dataAttributeRepo.Init(organisationId)

	da, err := dataAttributeRepo.Get(dataAttributeId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data attribute: %v", dataAttributeId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	var revisionResp revision.Revision

	if revisionId != "" {

		revisionResp, err = revision.GetByRevisionId(revisionId)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch revision: %v", dataAttributeId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

	} else {
		revisionResp, err = revision.GetLatestByDataAttributeId(dataAttributeId)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch revision: %v", dataAttributeId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
	}

	// Constructing the response
	var resp getDataAttributeResp
	resp.DataAttribute = da

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(revisionResp)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
