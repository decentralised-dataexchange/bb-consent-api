package onboard

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/user"
)

type validateUserEmailReq struct {
	Email string `valid:"required, email"`
}

type validateResp struct {
	Result  bool   `json:"result"` //True for valid email
	Message string `json:"message"`
}

// ValidateUserEmail Validates the user email
func ValidateUserEmail(w http.ResponseWriter, r *http.Request) {
	var validateReq validateUserEmailReq
	var valResp validateResp

	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &validateReq)

	valid, err := govalidator.ValidateStruct(validateReq)
	if valid != true {
		valResp.Result = false
		valResp.Message = err.Error()

		response, _ := json.Marshal(valResp)
		w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(response)
		return
	}

	valResp.Result = true
	valResp.Message = "Email address is valid and not in use in our system"

	sanitizedEmail := common.Sanitize(validateReq.Email)

	//Check whether the email is unique
	exist, err := user.EmailExist(sanitizedEmail)
	if err != nil {
		m := fmt.Sprintf("Failed to validate user email: %v", validateReq.Email)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	if exist == true {
		valResp.Result = false
		valResp.Message = "Email address is in use"
	}

	response, _ := json.Marshal(valResp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
