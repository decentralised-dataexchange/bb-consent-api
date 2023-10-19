package apikey

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/v2/apikey"
	"github.com/gorilla/mux"
)

type updateApiKeyReq struct {
	Apikey apikey.ApiKey `json:"apiKey" valid:"required"`
}

type updateApiKeyResp struct {
	Apikey apikey.ApiKey `json:"apiKey" valid:"required"`
}

// ConfigUpdateApiKey
func ConfigUpdateApiKey(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	apiKeyId := mux.Vars(r)[config.ApiKeyId]
	apiKeyId = common.Sanitize(apiKeyId)

	o, err := org.Get(organisationId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organisation: %v", organisationId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	organisationAdminId := o.Admins[0].UserID

	// Request body
	var apiKeyReq updateApiKeyReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &apiKeyReq)

	// Repository
	apiKeyRepo := apikey.ApiKeyRepository{}
	apiKeyRepo.Init(organisationId)
	toBeUpdatedApiKey, err := apiKeyRepo.Get(apiKeyId)
	if err != nil {
		m := fmt.Sprintf("Failed to remove api key:%v ", apiKeyId)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	expiryAt := int64(apiKeyReq.Apikey.ExpiryInDays * 24 * 60 * 60)
	if expiryAt <= 0 {
		//Default apikey expiry 1 month
		expiryAt = time.Now().Unix() + 60*60*24*30
		apiKeyReq.Apikey.ExpiryInDays = 30
	} else {
		expiryAt = time.Now().Unix() + int64(apiKeyReq.Apikey.ExpiryInDays)*60*60*24
	}

	currentApiKey, err := apikey.Create(apiKeyReq.Apikey.Scopes, expiryAt, organisationId, organisationAdminId)
	if err != nil {
		m := "Failed to create apiKey"
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	toBeUpdatedApiKey.Apikey = currentApiKey
	toBeUpdatedApiKey.ExpiryInDays = apiKeyReq.Apikey.ExpiryInDays
	toBeUpdatedApiKey.Scopes = apiKeyReq.Apikey.Scopes

	// Updates apikey
	savedApiKey, err := apiKeyRepo.Update(toBeUpdatedApiKey)
	if err != nil {
		m := fmt.Sprintf("Failed to remove apiKey:%v", apiKeyId)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := updateApiKeyResp{
		Apikey: savedApiKey,
	}
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
