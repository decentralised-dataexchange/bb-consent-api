package handlerv2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/actionlog"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/user"
)

type loginReq struct {
	Username string `valid:"required,email"`
	Password string `valid:"required"`
}

type loginResp struct {
	User  user.User
	Token iamToken
}

// LoginAdminUser Implements the admin users login
func LoginAdminUser(w http.ResponseWriter, r *http.Request) {
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
	accessToken, err := token.ParseToken(t.AccessToken)
	if err != nil {
		m := fmt.Sprintf("Failed to parse token for user:%v", lReq.Username)
		common.HandleError(w, status, m, err)
		return
	}

	u, err := user.GetByIamID(accessToken.IamID)
	if err != nil {
		m := fmt.Sprintf("User: %v does not exist", lReq.Username)
		common.HandleError(w, http.StatusUnauthorized, m, err)
		return
	}

	if len(u.Roles) == 0 {
		//Normal user can not login with this API.
		m := fmt.Sprintf("Non Admin User: %v tried admin login", lReq.Username)
		common.HandleError(w, http.StatusForbidden, m, err)
		return
	}

	actionLog := fmt.Sprintf("%v logged in", u.Email)
	actionlog.LogOrgSecurityCalls(u.ID.Hex(), u.Email, u.Roles[0].OrgID, actionLog)
	lResp := loginResp{u, t}
	resp, _ := json.Marshal(lResp)
	w.WriteHeader(http.StatusOK)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(resp)
}
