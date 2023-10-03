package handlerv2

import (
	"encoding/json"
	"net/http"

	"github.com/bb-consent/api/src/config"
)

// OrgListPolicy Handler to list all global policies
func OrgListPolicy(w http.ResponseWriter, r *http.Request) {

	// Constructing the response
	var resp globalPolicyConfigurationResp

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
