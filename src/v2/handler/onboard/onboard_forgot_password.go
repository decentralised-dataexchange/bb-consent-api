package onboard

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/user"
	"github.com/bb-consent/api/src/v2/iam"
)

type forgotPassword struct {
	Username string `json:"username" valid:"required,email"`
}

// OnboardForgotPassword User forgot the password, need to reset the password
func OnboardForgotPassword(w http.ResponseWriter, r *http.Request) {
	var fp forgotPassword

	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &fp)

	// validating request params
	valid, err := govalidator.ValidateStruct(fp)
	if !valid {
		log.Printf("Invalid request params for forgot password")
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	log.Printf("User: %v forgot password", fp.Username)

	sanitizedUserName := common.Sanitize(fp.Username)

	//Get user details from DB
	u, err := user.GetByEmail(sanitizedUserName)
	if err != nil {
		log.Printf("User with %v doesnt exist", fp.Username)
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}
	err = iam.ForgotPassword(u.IamID)
	if err != nil {
		m := "Failed to send email"
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
}
