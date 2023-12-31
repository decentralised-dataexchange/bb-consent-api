package onboard

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/otp"
	"github.com/bb-consent/api/internal/sms"
)

type verifyPhoneNumberReq struct {
	Phone string `valid:"required" json:"phone"`
}

// VerifyPhoneNumber Verifies the user phone number
func VerifyPhoneNumber(w http.ResponseWriter, r *http.Request) {
	verifyPhoneNumber(w, r, common.ClientTypeIos)
}

func generateVerificationCode() (code string, err error) {
	var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
	codeSize := 6
	b := make([]byte, codeSize)
	n, err := io.ReadAtLeast(rand.Reader, b, codeSize)
	if n != codeSize {
		return code, err
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b), nil
}

func sendPhoneVerificationMessage(msgTo string, message string) error {
	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + sms.TwilioConfig.AccountSid + "/Messages.json"

	// Pack up the data for our message
	msgData := url.Values{}

	// Add "+" before the phone number
	if !strings.Contains(msgTo, "+") {
		msgTo = "+" + msgTo
	}

	msgData.Set("To", msgTo)

	if strings.Contains(msgTo, "+1") {
		msgData.Set("From", "+15063065105")
	} else {
		msgData.Set("From", "+46769437629")
	}
	msgData.Set("Body", message)

	msgDataReader := *strings.NewReader(msgData.Encode())

	// Create HTTP request client
	client := &http.Client{}
	req, _ := http.NewRequest("POST", urlStr, &msgDataReader)
	req.SetBasicAuth(sms.TwilioConfig.AccountSid, sms.TwilioConfig.AuthToken)
	req.Header.Add("Accept", config.ContentTypeJSON)
	req.Header.Add(config.ContentTypeHeader, config.ContentTypeFormURLEncoded)

	// Make HTTP POST request and return message SID
	resp, _ := client.Do(req)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		defer resp.Body.Close()
		err := decoder.Decode(&data)
		if err == nil {
			fmt.Println(data["sid"])
		}
	} else {
		fmt.Println(resp.Status)
		return errors.New("Failed to send message")
	}
	return nil
}

// verifyPhoneNumber Verifies the user phone number
func verifyPhoneNumber(w http.ResponseWriter, r *http.Request, clientType int) {
	var verifyReq verifyPhoneNumberReq

	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &verifyReq)

	valid, err := govalidator.ValidateStruct(verifyReq)
	if valid != true {
		log.Printf("Invalid request params for verifying phone number")
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	vCode, err := generateVerificationCode()
	if err != nil {
		m := fmt.Sprintf("Failed to generate OTP :%v", verifyReq.Phone)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	var message strings.Builder
	message.Grow(32)
	if clientType == common.ClientTypeAndroid {
		fmt.Fprintf(&message, "[#]Thank you for signing up for iGrant.io! Your code is %s \n U1vUn/jAcoT", vCode)
	} else {
		fmt.Fprintf(&message, "Thank you for signing up for iGrant.io! Your code is %s", vCode)
	}

	err = sendPhoneVerificationMessage(verifyReq.Phone, message.String())
	if err != nil {
		m := fmt.Sprintf("Failed to send sms to :%v", verifyReq.Phone)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	var o otp.Otp
	o.Phone = verifyReq.Phone
	o.Otp = vCode

	sanitizedPhoneNumber := common.Sanitize(o.Phone)

	oldOtp, err := otp.SearchPhone(sanitizedPhoneNumber)
	if err == nil {
		otp.Delete(oldOtp.ID)
	}

	o, err = otp.Add(o)
	if err != nil {
		m := fmt.Sprintf("Failed to store otp details")
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
