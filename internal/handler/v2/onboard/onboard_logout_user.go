package onboard

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/actionlog"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/iam"
	"github.com/bb-consent/api/internal/individual"
	"github.com/bb-consent/api/internal/token"
	"github.com/bb-consent/api/internal/user"
)

// logoutUser
func logoutUser(accessToken string, refreshToken string, iamId string, organisationId string) error {

	orgAdmin, err := user.GetByIamID(iamId)
	if err != nil {
		// Repository
		individualRepo := individual.IndividualRepository{}
		individualRepo.Init(organisationId)

		_, err := individualRepo.GetByIamID(iamId)
		if err != nil {
			return err
		}

	}
	iam.LogoutUser(accessToken, refreshToken)
	// log security calls
	if len(orgAdmin.Roles) > 0 {
		actionLog := fmt.Sprintf("%v logged out", orgAdmin.Email)
		actionlog.LogOrgSecurityCalls(orgAdmin.ID, orgAdmin.Email, organisationId, actionLog)
	}
	return nil
}

type logoutReq struct {
	RefreshToken string `json:"refreshToken" valid:"required"`
}

// OnboardLogoutUser Logouts a user
func OnboardLogoutUser(w http.ResponseWriter, r *http.Request) {
	organisationId := r.Header.Get(config.OrganizationId)

	var lReq logoutReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &lReq)

	// validating request payload for logout
	valid, err := govalidator.ValidateStruct(lReq)
	if !valid {
		m := "Missing mandatory param refresh token"
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}
	headerType, accessToken, err := token.DecodeAuthHeader(r)

	if headerType != token.AuthorizationToken {
		m := "Failed to logout user"
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	iamId := token.GetIamID(r)

	// logout user
	logoutUser(accessToken, lReq.RefreshToken, iamId, organisationId)

	w.WriteHeader(http.StatusOK)
}
