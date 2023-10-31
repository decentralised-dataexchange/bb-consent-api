package onboard

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/iam"
	"github.com/bb-consent/api/internal/token"
)

type resetPasswordReq struct {
	CurrentPassword string `json:"currentPassword" valid:"required"`
	NewPassword     string `json:"newPassword" valid:"required"`
}

// ResetPassword Resets an user password
func OnboardResetPassword(w http.ResponseWriter, r *http.Request) {
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
	err = iam.ResetPassword(userIamID, resetReq.NewPassword)
	if err != nil {
		log.Printf("Failed to reset user:%v password ")
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
}
