package handlerv2

import (
	"encoding/json"
	"net/http"

	"github.com/bb-consent/api/src/config"
)

// UpdateGlobalPolicyConfigurationById Handler to update global policy configuration
func UpdateGlobalPolicyConfigurationById(w http.ResponseWriter, r *http.Request) {

	// Constructing the response
	var resp globalPolicyConfigurationResp

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
