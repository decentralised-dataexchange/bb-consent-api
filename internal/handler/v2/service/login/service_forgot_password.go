package login

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/iam"
	"github.com/bb-consent/api/internal/individual"
)

type forgotPassword struct {
	Username string `json:"username" valid:"required,email"`
}

// ServiceForgotPassword User forgot the password, need to reset the password
func ServiceForgotPassword(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)
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

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	//Get user details from DB
	u, err := individualRepo.GetIndividualByEmail(sanitizedUserName)
	if err != nil {
		log.Printf("User with %v doesnt exist", fp.Username)
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	err = iam.ForgotPassword(u.IamId)
	if err != nil {
		m := "Failed to send email"
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
}
