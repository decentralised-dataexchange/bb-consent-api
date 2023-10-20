package onboard

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/iam"
)

type tokenReq struct {
	RefreshToken string `valid:"required" json:"refreshToken"`
	ClientID     string `valid:"required" json:"clientId"`
}

type refreshTokenResp struct {
	AccessToken      string `json:"accessToken"`
	ExpiresIn        int    `json:"expiresIn"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
	RefreshToken     string `json:"refreshToken"`
	TokenType        string `json:"tokenType"`
}

// OnboardRefreshToken
func OnboardRefreshToken(w http.ResponseWriter, r *http.Request) {
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

	client := iam.GetClient()

	t, err := iam.RefreshToken(tReq.ClientID, tReq.RefreshToken, iam.IamConfig.Realm, client)
	if err != nil {
		m := "failed to get token from refresh token"
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}
	resp := refreshTokenResp{
		AccessToken:      t.AccessToken,
		ExpiresIn:        t.ExpiresIn,
		RefreshExpiresIn: t.RefreshExpiresIn,
		RefreshToken:     t.RefreshToken,
		TokenType:        t.TokenType,
	}
	response, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}
