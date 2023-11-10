package onboard

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/iam"
	"github.com/bb-consent/api/internal/individual"
	"github.com/bb-consent/api/internal/token"
	"github.com/bb-consent/api/internal/user"
)

type resetPasswordReq struct {
	CurrentPassword string `json:"currentPassword" valid:"required"`
	NewPassword     string `json:"newPassword" valid:"required"`
}

// ResetPassword Resets an user password
func OnboardResetPassword(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	userIamID := token.GetIamID(r)

	var resetReq resetPasswordReq
	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &resetReq)

	valid, err := govalidator.ValidateStruct(resetReq)
	if !valid {
		log.Printf("Missing mandatory params required to reset password")
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	//Get user details from DB
	var email string
	u, err := individualRepo.GetByIamID(userIamID)
	if err != nil {
		u, err := user.GetByIamID(userIamID)
		if err != nil {
			m := "Failed to fetch user"
			common.HandleErrorV2(w, http.StatusBadRequest, m, err)
			return
		}
		email = u.Email
	} else {
		email = u.Email
	}

	// reset user password
	err = iam.ResetPassword(userIamID, email, resetReq.CurrentPassword, resetReq.NewPassword)
	if err != nil {
		m := fmt.Sprintf("Failed to reset user:%v password", userIamID)
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
}
