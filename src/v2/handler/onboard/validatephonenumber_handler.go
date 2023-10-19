package onboard

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/otp"
	"github.com/bb-consent/api/src/user"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type validatePhoneNumberReq struct {
	Phone string `valid:"required" json:"phone"`
}

// ValidatePhoneNumber Check if the phone number is already in use
func ValidatePhoneNumber(w http.ResponseWriter, r *http.Request) {
	var validateReq validatePhoneNumberReq
	var valResp validateResp

	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &validateReq)

	// validating request payload
	valid, err := govalidator.ValidateStruct(validateReq)
	if !valid {
		log.Printf("Missing mandatory params for validating phone number")
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	valResp.Result = true
	valResp.Message = "Phone number is not in use"

	sanitizedPhoneNumber := common.Sanitize(validateReq.Phone)

	//Check whether the phone number is unique
	exist, err := user.PhoneNumberExist(sanitizedPhoneNumber)
	if err != nil {
		m := fmt.Sprintf("Failed to validate user phone number: %v", validateReq.Phone)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	if exist {
		valResp.Result = false
		valResp.Message = "Phone number is in use"
		response, _ := json.Marshal(valResp)
		w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		w.Write(response)
		return
	}

	//Check whether the phone number is in otp colleciton
	o, err := otp.PhoneNumberExist(sanitizedPhoneNumber)
	if err != nil {
		m := fmt.Sprintf("Failed to find otp for phone number: %v", err)
		log.Println(m)
	}

	if o != (otp.Otp{}) {
		if primitive.NewObjectID().Timestamp().Sub(o.ID.Timestamp()) > 2*time.Minute {
			err = otp.Delete(o.ID.Hex())
			if err != nil {
				m := "Failed to clear expired otp"
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}
		} else {
			valResp.Result = false
			valResp.Message = "Phone number is in use"
		}
	}

	response, _ := json.Marshal(valResp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
