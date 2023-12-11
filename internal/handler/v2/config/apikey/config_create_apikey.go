package apikey

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/apikey"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/org"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type addApiKeyReq struct {
	Apikey apikey.ApiKey `json:"apiKey" valid:"required"`
}

type addApiKeyResp struct {
	Apikey apikey.ApiKey `json:"apiKey" valid:"required"`
}

// ConfigCreateApiKey
func ConfigCreateApiKey(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	o, err := org.Get(organisationId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organisation: %v", organisationId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	organisationAdminId := o.Admins[0].UserID

	// Request body
	var apiKeyReq addApiKeyReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &apiKeyReq)

	// validating request payload
	valid, err := govalidator.ValidateStruct(apiKeyReq)
	if !valid {
		m := "missing mandatory params for creating api key"
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	// validate scopes
	validScopes := apikey.ValidateScopes(apiKeyReq.Apikey.Scopes)
	if !validScopes {
		m := "Invalid scopes provided for creating api key"
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	// Repository
	apiKeyRepo := apikey.ApiKeyRepository{}
	apiKeyRepo.Init(organisationId)

	expiryAt := int64(apiKeyReq.Apikey.ExpiryInDays)
	if expiryAt <= 0 {
		//Default apikey expiry 1 month
		expiryAt = time.Now().Unix() + 60*60*24*30
		apiKeyReq.Apikey.ExpiryInDays = 30
	} else {
		expiryAt = time.Now().Unix() + int64(apiKeyReq.Apikey.ExpiryInDays)*60*60*24
	}

	key, err := apikey.Create(apiKeyReq.Apikey.Scopes, expiryAt, organisationId, organisationAdminId)
	if err != nil {
		m := "Failed to create apiKey"
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	// Convert the int timestamp to a time.Time value
	expiryTime := time.Unix(expiryAt, 0)

	expiryTimestamp := expiryTime.UTC().Format("2006-01-02T15:04:05Z")

	var newApiKey apikey.ApiKey
	newApiKey.Id = primitive.NewObjectID().Hex()
	newApiKey.Name = apiKeyReq.Apikey.Name
	newApiKey.Scopes = apiKeyReq.Apikey.Scopes
	newApiKey.Apikey = key
	newApiKey.ExpiryInDays = apiKeyReq.Apikey.ExpiryInDays
	newApiKey.OrganisationId = organisationId
	newApiKey.IsDeleted = false
	newApiKey.ExpiryTimestamp = expiryTimestamp

	apiKey, err := apiKeyRepo.Add(newApiKey)
	if err != nil {
		m := fmt.Sprintf("Failed to add api key: %v", organisationId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := addApiKeyResp{
		Apikey: apiKey,
	}
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
