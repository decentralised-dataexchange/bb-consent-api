package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/user"
)

type tokenResp struct {
	AccessToken      string `json:"accessToken"`
	ExpiresIn        int    `json:"expiresIn"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
	RefreshToken     string `json:"refreshToken"`
	TokenType        string `json:"tokenType"`
}

type userLoginResp struct {
	Individual user.UserV2 `json:"individual"`
	Token      tokenResp   `json:"token"`
}

// LoginUser Implements the user login
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var lReq loginReq

	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &lReq)

	log.Printf("Login username: %v", lReq.Username)

	// validating the request payload
	valid, err := govalidator.ValidateStruct(lReq)

	if !valid {
		log.Printf("Invalid request params for authentication")
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	t, status, iamErr, err := getToken(lReq.Username, lReq.Password, "igrant-ios-app", iamConfig.Realm)
	if err != nil {
		if (iamError{}) != iamErr {
			resp, _ := json.Marshal(iamErr)
			w.WriteHeader(status)
			w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
			w.Write(resp)
			return
		}
		m := fmt.Sprintf("Failed to get token for user:%v", lReq.Username)
		common.HandleErrorV2(w, status, m, err)
		return
	}

	accessToken, err := token.ParseToken(t.AccessToken)
	if err != nil {
		m := fmt.Sprintf("Failed to parse token for user:%v", lReq.Username)
		common.HandleErrorV2(w, status, m, err)
		return
	}
	u, err := user.GetByIamIDV2(accessToken.IamID)
	if err != nil {
		m := fmt.Sprintf("User: %v does not exist", lReq.Username)
		common.HandleErrorV2(w, status, m, err)
		return
	}
	tResp := tokenResp{
		AccessToken:      t.AccessToken,
		ExpiresIn:        t.ExpiresIn,
		RefreshExpiresIn: t.RefreshExpiresIn,
		RefreshToken:     t.RefreshToken,
		TokenType:        t.TokenType,
	}

	lResp := userLoginResp{u, tResp}
	resp, _ := json.Marshal(lResp)
	w.WriteHeader(http.StatusOK)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(resp)

}
