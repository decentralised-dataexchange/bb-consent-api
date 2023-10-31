package apikey

import (
	"encoding/json"
	"net/http"

	"github.com/bb-consent/api/internal/apikey"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/gorilla/mux"
)

type deleteApiKeyResp struct {
	Apikey apikey.ApiKey `json:"apiKey" valid:"required"`
}

// ConfigDeleteAPIKey
func ConfigDeleteApiKey(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	apiKeyId := mux.Vars(r)[config.ApiKeyId]
	apiKeyId = common.Sanitize(apiKeyId)

	// Repository
	apiKeyRepo := apikey.ApiKeyRepository{}
	apiKeyRepo.Init(organisationId)
	apiKey, err := apiKeyRepo.Get(apiKeyId)
	if err != nil {
		m := "Failed to remove api key for organisation"
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	apiKey.IsDeleted = true

	// Deletes api key
	apiKey, err = apiKeyRepo.Update(apiKey)
	if err != nil {
		m := "Failed to remove api key for organisation "
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	resp := deleteApiKeyResp{
		Apikey: apiKey,
	}
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
