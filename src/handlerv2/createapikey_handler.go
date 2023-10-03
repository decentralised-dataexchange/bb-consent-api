package handlerv2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/apikey"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/user"
)

type apiKeyResponse struct {
	User   string
	APIKey string
}

// CreateAPIKey Create the API key for the user
func CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := token.GetUserID(r)

	key, err := apikey.Create(userID)
	if err != nil {
		m := fmt.Sprintf("Failed to create apiKey for user:%v err:%v", token.GetUserName(r), err)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	err = user.UpdateAPIKey(userID, key)
	if err != nil {
		m := fmt.Sprintf("Failed to store apiKey for user:%v err:%v", token.GetUserName(r), err)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	apiKeyResp := apiKeyResponse{token.GetUserName(r), key}
	response, _ := json.Marshal(apiKeyResp)

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
