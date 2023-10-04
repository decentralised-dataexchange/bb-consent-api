package handlerv2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/user"
)

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
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
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
		common.HandleError(w, status, m, err)
		return
	}
	sanitizedUserName := common.Sanitize(lReq.Username)

	//TODO: Remove me when the auth server is per dev environment
	u, err := user.GetByEmail(sanitizedUserName)
	if err != nil {
		m := fmt.Sprintf("Login failed for non existant user:%v", lReq.Username)
		common.HandleError(w, http.StatusUnauthorized, m, err)
		return
	}

	if len(u.Roles) > 0 {
		m := fmt.Sprintf("Login not allowed for admin users:%v", lReq.Username)
		common.HandleError(w, http.StatusUnauthorized, m, err)
		return
	}

	resp, _ := json.Marshal(t)
	w.WriteHeader(http.StatusOK)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(resp)
}
