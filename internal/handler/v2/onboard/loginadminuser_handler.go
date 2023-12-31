package onboard

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/actionlog"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/fixture"
	"github.com/bb-consent/api/internal/iam"
	"github.com/bb-consent/api/internal/token"
	"github.com/bb-consent/api/internal/user"
)

type loginReq struct {
	Username string `json:"username" valid:"required,email"`
	Password string `json:"password" valid:"required"`
}

type loginResp struct {
	AccessToken      string `json:"accessToken"`
	ExpiresIn        int    `json:"expiresIn"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
	RefreshToken     string `json:"refreshToken"`
	TokenType        string `json:"tokenType"`
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
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	client := iam.GetClient()

	t, err := iam.GetToken(lReq.Username, lReq.Password, iam.IamConfig.Realm, client)
	if err != nil {
		m := fmt.Sprintf("Failed to get token for user:%v", lReq.Username)
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}
	accessToken, err := token.ParseToken(t.AccessToken)
	if err != nil {
		m := fmt.Sprintf("Failed to parse token for user:%v", lReq.Username)
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	u, err := user.GetByIamID(accessToken.IamID)
	if err != nil {
		m := fmt.Sprintf("User: %v does not exist", lReq.Username)
		common.HandleErrorV2(w, http.StatusUnauthorized, m, err)
		return
	}

	if len(u.Roles) == 0 {
		//Normal user can not login with this API.
		m := fmt.Sprintf("Non Admin User: %v tried admin login", lReq.Username)
		common.HandleErrorV2(w, http.StatusForbidden, m, err)
		return
	}

	// Load default user image
	u, err = fixture.LoadOrganisationAdminAvatarImageAssets(u, r.Host)
	if err != nil {
		common.HandleErrorV2(w, http.StatusUnauthorized, "Failed to load default avatar image for admin", err)
		return
	}

	actionLog := fmt.Sprintf("%v logged in", u.Email)
	actionlog.LogOrgSecurityCalls(u.ID, u.Email, u.Roles[0].OrgID, actionLog)
	lResp := loginResp{
		AccessToken:      t.AccessToken,
		ExpiresIn:        t.ExpiresIn,
		RefreshExpiresIn: t.RefreshExpiresIn,
		RefreshToken:     t.RefreshToken,
		TokenType:        t.TokenType,
	}
	resp, _ := json.Marshal(lResp)
	w.WriteHeader(http.StatusOK)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(resp)
}
