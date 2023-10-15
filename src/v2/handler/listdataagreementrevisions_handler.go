package handler

import (
	"encoding/json"
	"net/http"

	"github.com/bb-consent/api/src/config"
)

// ListDataAgreementRevisions Gets an organization data agreements revisions
func ListDataAgreementRevisions(w http.ResponseWriter, r *http.Request) {

	response, _ := json.Marshal(getPurposesResp{})
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
