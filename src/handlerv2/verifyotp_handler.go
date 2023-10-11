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
	"github.com/bb-consent/api/src/otp"
)

type verifyOtpReq struct {
	Phone string `valid:"required" json:"phone"`
	Otp   string `valid:"required" json:"otp"`
}

// VerifyOtp Verifies the Otp
func VerifyOtp(w http.ResponseWriter, r *http.Request) {
	var otpReq verifyOtpReq
	var valResp validateResp

	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &otpReq)

	valid, err := govalidator.ValidateStruct(otpReq)
	if valid != true {
		log.Printf("Missing mandatory params for verify otp")
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	sanitizedPhoneNumber := common.Sanitize(otpReq.Phone)

	o, err := otp.SearchPhone(sanitizedPhoneNumber)
	if err != nil {
		valResp.Result = false
		valResp.Message = "Unregistered phone number: " + otpReq.Phone
		response, _ := json.Marshal(valResp)
		w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(response)
		return
	}

	valResp.Result = true
	valResp.Message = "Otp validatiation Succeeded"
	if err != nil || o.Otp != otpReq.Otp || o.Phone != otpReq.Phone {
		valResp.Result = false
		valResp.Message = "Otp validatiation failed with mismatch in otp data"

	} else {
		o.Verified = true
		//TODO: When user registration comes, locate the details and match and then remove this entry
		//TODO: Periodic delete of stale OTP entries based on creation time needed
		err := otp.UpdateVerified(o)
		if err != nil {
			m := fmt.Sprintf("Failed to update internal database")
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
	}

	response, _ := json.Marshal(valResp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
