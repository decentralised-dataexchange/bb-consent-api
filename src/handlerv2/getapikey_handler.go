package handlerv2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/user"
)

// GetAPIKey Get the API key from user
func GetAPIKey(w http.ResponseWriter, r *http.Request) {
	userID := token.GetUserID(r)

	apiKey, err := user.GetAPIKey(userID)
	if err != nil {
		m := fmt.Sprintf("Failed to get apiKey for user:%v err:%v", token.GetUserName(r), err)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	apiKeyResp := apiKeyResponse{token.GetUserName(r), apiKey}
	response, _ := json.Marshal(apiKeyResp)

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
