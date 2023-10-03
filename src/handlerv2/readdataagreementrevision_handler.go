package handlerv2

import (
	"encoding/json"
	"net/http"

	"github.com/bb-consent/api/src/config"
)

// ReadDataAgreementRevision Gets an organization data agreements revision
func ReadDataAgreementRevision(w http.ResponseWriter, r *http.Request) {

	response, _ := json.Marshal(getPurposesResp{})
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
