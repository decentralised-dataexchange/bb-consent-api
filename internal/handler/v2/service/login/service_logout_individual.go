package login

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/iam"
	"github.com/bb-consent/api/internal/individual"
	"github.com/bb-consent/api/internal/token"
)

// logoutUser
func logoutUser(accessToken string, refreshToken string, iamId string, organisationId string) error {

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	_, err := individualRepo.GetByIamID(iamId)
	if err != nil {
		return err
	}
	iam.LogoutUser(accessToken, refreshToken)
	return nil
}

type logoutReq struct {
	RefreshToken string `json:"refreshToken" valid:"required"`
}

// ServiceLogoutUser Logouts a user
func ServiceLogoutUser(w http.ResponseWriter, r *http.Request) {
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

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
