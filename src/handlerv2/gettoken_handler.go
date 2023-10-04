package handlerv2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
)

type tokenReq struct {
	RefreshToken string `valid:"required"`
	ClientID     string `valid:"required"`
}

// GetToken return access token when refresh token is given
func GetToken(w http.ResponseWriter, r *http.Request) {
	var tReq tokenReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &tReq)

	// validating request payload for refreshing tokens
	valid, err := govalidator.ValidateStruct(tReq)

	if !valid {
		log.Printf("Failed to refresh token")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	data := url.Values{}
	data.Set("refresh_token", tReq.RefreshToken)
	data.Add("client_id", tReq.ClientID)
	data.Add("grant_type", "refresh_token")

	resp, err := http.PostForm(iamConfig.URL+"/realms/"+iamConfig.Realm+"/protocol/openid-connect/token", data)
	if err != nil {
		//m := fmt.Sprintf("Failed to get token from refresh token for user:%v", token.GetUserName(r))
		m := fmt.Sprintf("Failed to get token from refresh token")
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//m := fmt.Sprintf("Failed to get token from refresh token user:%v", token.GetUserName(r))
		m := fmt.Sprintf("Failed to get token from refresh token")
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		var e iamError
		json.Unmarshal(body, &e)
		response, _ := json.Marshal(e)
		w.WriteHeader(resp.StatusCode)
		w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
		w.Write(response)
		return
	}

	var tok iamToken
	json.Unmarshal(body, &tok)
	response, _ := json.Marshal(tok)
	w.WriteHeader(resp.StatusCode)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}
