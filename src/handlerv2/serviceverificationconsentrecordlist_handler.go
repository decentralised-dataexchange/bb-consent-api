package handlerv2

import (
	"net/http"

	"github.com/bb-consent/api/src/config"
)

func ServiceVerificationConsentRecordList(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)

}
