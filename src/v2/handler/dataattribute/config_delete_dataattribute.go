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

// ConfigDeleteDataAttribute
func ConfigDeleteDataAttribute(w http.ResponseWriter, r *http.Request) {
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)
	dataAttributeId := mux.Vars(r)[config.DataAttributeId]
	dataAttributeId = common.Sanitize(dataAttributeId)

	// Repository
	dataAttributeRepo := dataattribute.DataAttributeRepository{}
	dataAttributeRepo.Init(organisationId)

	currentDataAttribute, err := dataAttributeRepo.Get(dataAttributeId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data attribute: %v", dataAttributeId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	currentRevision, err := revision.GetLatestByDataAttributeId(dataAttributeId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch revisions: %v", dataAttributeId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	currentDataAttribute.IsDeleted = true

	_, err = dataAttributeRepo.Update(currentDataAttribute)
	if err != nil {
		m := fmt.Sprintf("Failed to delete data attribute: %v", dataAttributeId)
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
