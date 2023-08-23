package handler

import (
	"bytes"
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
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/actionlog"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/email"
	"github.com/bb-consent/api/src/otp"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/user"
	"github.com/globalsign/mgo/bson"
)

type registerReq struct {
	Username    string `valid:"required,email"` //Username is email
	Name        string
	Password    string `valid:"required,length(8|16)"`
	Phone       string `valid:"required"`
	IsAdmin     bool
	HlcRegister bool
}

type iamCredentials struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type iamUserRegisterReq struct {
	Username    string           `json:"username"`
	Firstname   string           `json:"firstName"`
	Lastname    string           `json:"lastName"`
	Email       string           `json:"email"`
	Enabled     bool             `json:"enabled"`
	RealmRoles  []string         `json:"realmRoles"`
	Credentials []iamCredentials `json:"credentials"`
}

type iamToken struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
}

type iamError struct {
	ErrorType string `json:"error"`
	Error     string `json:"error_description"`
}

var timeout time.Duration

var iamConfig config.Iam
var twilioConfig config.Twilio

// IamInit Initialize the IAM handler
func IamInit(config *config.Configuration) {
	iamConfig = config.Iam
	twilioConfig = config.Twilio
	timeout = time.Duration(time.Duration(iamConfig.Timeout) * time.Second)

	/*
		memStorage := storage.NewMemoryStorage()
		s := scheduler.New(memStorage)
		_, err := s.RunEvery(24*time.Hour, clearStaleOtps)
		if err != nil {
			log.Printf("err in scheduling clearStaleOtps: %v", err)
		}

		//TODO: Enable this later phase
		//s.Start()
	*/
}

// RegisterUser Registers a new user
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var regReq registerReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &regReq)

	valid, err := govalidator.ValidateStruct(regReq)
	if !valid {
		handleError(w, regReq.Username, http.StatusBadRequest, iamError{"Invalid user input", err.Error()}, err)
		return
	}

	var iamRegReq iamUserRegisterReq
	iamRegReq.Username = regReq.Username
	iamRegReq.Firstname = regReq.Name
	//iamRegReq.Lastname = regReq.Lastname
	iamRegReq.Email = regReq.Username
	iamRegReq.Enabled = true

	iamRegReq.Credentials = append(iamRegReq.Credentials, iamCredentials{"password", regReq.Password})
	iamRegReq.RealmRoles = append(iamRegReq.RealmRoles, "organization-admin")

	t, status, iamErr, err := getAdminToken()
	if err != nil {
		log.Printf("Failed to get admin token, user: %v registration", regReq.Username)
		handleError(w, regReq.Username, status, iamErr, err)
		return
	}

	status, iamErr, err = registerUser(iamRegReq, t.AccessToken)
	if err != nil {
		log.Printf("Failed to register user: %v err: %v", regReq.Username, err)
		handleError(w, regReq.Username, status, iamErr, err)
		return
	}
	userIamID, status, iamErr, err := getUserIamID(regReq.Username, t.AccessToken)
	if err != nil {
		log.Printf("Failed to get userID for user: %v err: %v", regReq.Username, err)
		handleError(w, regReq.Username, status, iamErr, err)
		return
	}
	if regReq.IsAdmin {
		status, iamErr, err = setAdminRole(userIamID, t.AccessToken)
		if err != nil {
			log.Printf("Failed to set roles for user: %v iam id: %v err: %v", regReq.Username, userIamID, err)
			handleError(w, regReq.Username, status, iamErr, err)
			return
		}
	}
	var u user.User
	u.Name = regReq.Name
	u.IamID = userIamID
	u.Email = regReq.Username
	u.Phone = regReq.Phone

	u, err = user.Add(u)
	if err != nil {
		log.Printf("Failed to add user: %v id: %v to Db err: %v", regReq.Username, userIamID, err)
		m := fmt.Sprintf("Failed to register user: %v", regReq.Username)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// Sending welcome email
	go email.SendWelcomeEmail(u.Email, u.Name, "Welcome to iGrant.io", "", email.SMTPConfig.AdminEmail)

	log.Printf("successfully registered user: %v", regReq.Username)
	//TODO; json response needed for the creation successful.
	type createResponse struct {
		Msg string `json:"msg"`
	}
	resp := createResponse{"User created successfully"}
	response, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

func handleError(w http.ResponseWriter, userName string, status int, iamErr iamError, err error) {
	if (iamError{}) != iamErr {
		log.Printf("Failed to register err:%v", err)
		resp, _ := json.Marshal(iamErr)
		w.WriteHeader(status)
		w.Header().Set("Content-Type", "application/json")
		w.Write(resp)
		return
	}
	m := fmt.Sprintf("Failed to register user:%v", userName)
	common.HandleError(w, status, m, err)
	return
}

func getAdminToken() (iamToken, int, iamError, error) {
	t, status, iamErr, err := getToken(iamConfig.AdminUser, iamConfig.AdminPassword, "admin-cli", "master")
	return t, status, iamErr, err
}

// registerUser Registers a new user
func registerUser(iamRegReq iamUserRegisterReq, adminToken string) (int, iamError, error) {
	var e iamError
	var status = http.StatusInternalServerError
	jsonReq, _ := json.Marshal(iamRegReq)
	req, err := http.NewRequest("POST", iamConfig.URL+"/auth/admin/realms/"+iamConfig.Realm+"/users", bytes.NewBuffer(jsonReq))
	if err != nil {
		return status, e, err
	}

	req.Header.Add("Authorization", "Bearer "+adminToken)
	req.Header.Add("Content-Type", "application/json")

	//dump, err := httputil.DumpRequest(req, true)
	//dump, err := httputil.DumpRequestOut(req, true)
	//log.Printf("\n %q \n", dump)

	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return status, e, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		type errorMsg struct {
			ErrorMessage string `json:"errorMessage"`
		}
		var errMsg errorMsg
		json.Unmarshal(body, &errMsg)
		e.Error = errMsg.ErrorMessage
		e.ErrorType = "Creation failed"
		return resp.StatusCode, e, errors.New("failed to register user")
	}
	return resp.StatusCode, e, err
}

func getUserIamID(userName string, adminToken string) (string, int, iamError, error) {
	var userID = ""
	var status = http.StatusInternalServerError
	var e iamError
	req, err := http.NewRequest("GET", iamConfig.URL+"/auth/admin/realms/"+iamConfig.Realm+"/users"+"?username="+userName, nil)
	if err != nil {
		return userID, status, e, err
	}
	//log.Printf("token: %v", t)
	req.Header.Add("Authorization", "Bearer "+adminToken)
	req.Header.Add("Content-Type", "application/json")

	//dump, err := httputil.DumpRequest(req, true)
	//dump, err := httputil.DumpRequestOut(req, true)
	//log.Printf("\n %q \n", dump)

	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return userID, status, e, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		json.Unmarshal(body, &e)
		return userID, resp.StatusCode, e, errors.New("failed to register user")
	}
	type userDetails struct {
		ID string `json:"id"`
	}
	var users []userDetails
	json.Unmarshal(body, &users)
	return users[0].ID, resp.StatusCode, e, err
}

type rReq struct {
	ClientRole         bool   `json:"clientRole"`
	Composite          bool   `json:"composite"`
	ContainerID        string `json:"containerId"`
	Description        string `json:"description"`
	ID                 string `json:"id"`
	Name               string `json:"name"`
	ScopeParamRequired bool   `json:"scopeParamRequired"`
}

// TODO: Get this from the IAM endpoint
var realmRoleOrgAdmin = "70698dc8-3202-4668-a982-4d95107399d4"

func setAdminRole(userID string, adminToken string) (int, iamError, error) {
	var status = http.StatusInternalServerError
	var e iamError

	var iReq []rReq
	iReq = append(iReq, rReq{false, false, iamConfig.Realm, "${organization.admin}", realmRoleOrgAdmin, "organization-admin", false})
	jsonReq, _ := json.Marshal(iReq)
	req, err := http.NewRequest("POST", iamConfig.URL+"/auth/admin/realms/"+iamConfig.Realm+"/users/"+userID+"/role-mappings/realm", bytes.NewBuffer(jsonReq))
	if err != nil {
		return status, e, err
	}

	req.Header.Add("Authorization", "Bearer "+adminToken)
	req.Header.Add("Content-Type", "application/json")

	//dump, err := httputil.DumpRequest(req, true)
	//dump, err := httputil.DumpRequestOut(req, true)
	//fmt.Printf("\n %q \n", dump)

	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to set user roles: with status:%v", resp.StatusCode)
		return status, e, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	json.Unmarshal(body, &e)
	return resp.StatusCode, e, err
}

func getToken(username string, password string, clientID string, realm string) (iamToken, int, iamError, error) {
	var tok iamToken
	var e iamError
	var status = http.StatusInternalServerError

	data := url.Values{}
	data.Set("username", username)
	data.Add("password", password)
	data.Add("client_id", clientID)
	data.Add("grant_type", "password")

	resp, err := http.PostForm(iamConfig.URL+"/auth/realms/"+realm+"/protocol/openid-connect/token", data)
	if err != nil {
		return tok, status, e, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return tok, status, e, err
	}
	if resp.StatusCode != http.StatusOK {
		var e iamError
		json.Unmarshal(body, &e)
		return tok, resp.StatusCode, e, errors.New("failed to get token")
	}
	json.Unmarshal(body, &tok)

	return tok, resp.StatusCode, e, err
}

type loginReq struct {
	Username string `valid:"required,email"`
	Password string `valid:"required"`
}

type loginResp struct {
	User  user.User
	Token iamToken
}

// LoginUser Implements the user login
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var lReq loginReq

	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &lReq)

	log.Printf("Login username: %v", lReq.Username)

	// validating the request payload
	valid, err := govalidator.ValidateStruct(lReq)

	if !valid {
		log.Printf("Invalid request params for authentication")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	t, status, iamErr, err := getToken(lReq.Username, lReq.Password, "igrant-ios-app", iamConfig.Realm)
	if err != nil {
		if (iamError{}) != iamErr {
			resp, _ := json.Marshal(iamErr)
			w.WriteHeader(status)
			w.Header().Set("Content-Type", "application/json")
			w.Write(resp)
			return
		}
		m := fmt.Sprintf("Failed to get token for user:%v", lReq.Username)
		common.HandleError(w, status, m, err)
		return
	}
	//TODO: Remove me when the auth server is per dev environment
	u, err := user.GetByEmail(lReq.Username)
	if err != nil {
		m := fmt.Sprintf("Login failed for non existant user:%v", lReq.Username)
		common.HandleError(w, http.StatusUnauthorized, m, err)
		return
	}

	if len(u.Roles) > 0 {
		m := fmt.Sprintf("Login not allowed for admin users:%v", lReq.Username)
		common.HandleError(w, http.StatusUnauthorized, m, err)
		return
	}

	resp, _ := json.Marshal(t)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// LoginUserV11 Implements the user login V1.1
func LoginUserV11(w http.ResponseWriter, r *http.Request) {
	var lReq loginReq

	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &lReq)

	log.Printf("Login username: %v", lReq.Username)

	// validating the request payload
	valid, err := govalidator.ValidateStruct(lReq)

	if !valid {
		log.Printf("Invalid request params for authentication")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	//TODO: Check whether user exist in the user db before returning token

	t, status, iamErr, err := getToken(lReq.Username, lReq.Password, "igrant-ios-app", iamConfig.Realm)
	if err != nil {
		if (iamError{}) != iamErr {
			resp, _ := json.Marshal(iamErr)
			w.WriteHeader(status)
			w.Header().Set("Content-Type", "application/json")
			w.Write(resp)
			return
		}
		m := fmt.Sprintf("Failed to get token for user:%v", lReq.Username)
		common.HandleError(w, status, m, err)
		return
	}

	accessToken, err := token.ParseToken(t.AccessToken)
	if err != nil {
		m := fmt.Sprintf("Failed to parse token for user:%v", lReq.Username)
		common.HandleError(w, status, m, err)
		return
	}
	u, err := user.GetByIamID(accessToken.IamID)
	if err != nil {
		m := fmt.Sprintf("User: %v does not exist", lReq.Username)
		common.HandleError(w, status, m, err)
		return
	}

	lResp := loginResp{u, t}
	resp, _ := json.Marshal(lResp)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

// LoginAdminUser Implements the admin users login
func LoginAdminUser(w http.ResponseWriter, r *http.Request) {
	var lReq loginReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &lReq)

	log.Printf("Login username: %v", lReq.Username)

	// validating the request payload
	valid, err := govalidator.ValidateStruct(lReq)

	if !valid {
		log.Printf("Invalid request params for authentication")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	t, status, iamErr, err := getToken(lReq.Username, lReq.Password, "igrant-ios-app", iamConfig.Realm)
	if err != nil {
		if (iamError{}) != iamErr {
			resp, _ := json.Marshal(iamErr)
			w.WriteHeader(status)
			w.Header().Set("Content-Type", "application/json")
			w.Write(resp)
			return
		}
		m := fmt.Sprintf("Failed to get token for user:%v", lReq.Username)
		common.HandleError(w, status, m, err)
		return
	}
	accessToken, err := token.ParseToken(t.AccessToken)
	if err != nil {
		m := fmt.Sprintf("Failed to parse token for user:%v", lReq.Username)
		common.HandleError(w, status, m, err)
		return
	}

	u, err := user.GetByIamID(accessToken.IamID)
	if err != nil {
		m := fmt.Sprintf("User: %v does not exist", lReq.Username)
		common.HandleError(w, http.StatusUnauthorized, m, err)
		return
	}

	if len(u.Roles) == 0 {
		//Normal user can not login with this API.
		m := fmt.Sprintf("Non Admin User: %v tried admin login", lReq.Username)
		common.HandleError(w, http.StatusForbidden, m, err)
		return
	}

	actionLog := fmt.Sprintf("%v logged in", u.Email)
	actionlog.LogOrgSecurityCalls(u.ID.Hex(), u.Email, u.Roles[0].OrgID, actionLog)
	lResp := loginResp{u, t}
	resp, _ := json.Marshal(lResp)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

type validateUserEmailReq struct {
	Email string `valid:"required, email"`
}

type validateResp struct {
	Result  bool //True for valid email
	Message string
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
		return
	}

	valResp.Result = true
	valResp.Message = "Email address is valid and not in use in our system"

	//Check whether the email is unique
	exist, err := user.EmailExist(validateReq.Email)
	if err != nil {
		m := fmt.Sprintf("Failed to validate user email: %v", validateReq.Email)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	if exist == true {
		valResp.Result = false
		valResp.Message = "Email address is in use"
	}

	response, _ := json.Marshal(valResp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

type validatePhoneNumberReq struct {
	Phone string `valid:"required"`
}

// ValidatePhoneNumber Check if the phone number is already in use
func ValidatePhoneNumber(w http.ResponseWriter, r *http.Request) {
	var validateReq validatePhoneNumberReq
	var valResp validateResp

	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &validateReq)

	// validating request payload
	valid, err := govalidator.ValidateStruct(validateReq)
	if valid != true {
		log.Printf("Missing mandatory params for validating phone number")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	valResp.Result = true
	valResp.Message = "Phone number is not in use"

	//Check whether the phone number is unique
	exist, err := user.PhoneNumberExist(validateReq.Phone)
	if err != nil {
		m := fmt.Sprintf("Failed to validate user phone number: %v", validateReq.Phone)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	if exist == true {
		valResp.Result = false
		valResp.Message = "Phone number is in use"
		response, _ := json.Marshal(valResp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
		return
	}

	//Check whether the phone number is in otp colleciton
	o, err := otp.PhoneNumberExist(validateReq.Phone)
	if err != nil {
		m := fmt.Sprintf("Failed to validate user phone number: %v", validateReq.Phone)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	if o != (otp.Otp{}) {
		if bson.NewObjectId().Time().Sub(o.ID.Time()) > 2*time.Minute {
			err = otp.Delete(o.ID.Hex())
			if err != nil {
				m := fmt.Sprintf("Failed to clear expired otp")
				common.HandleError(w, http.StatusInternalServerError, m, err)
				return
			}
		} else {
			valResp.Result = false
			valResp.Message = "Phone number is in use"
		}
	}

	response, _ := json.Marshal(valResp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

type verifyPhoneNumberReq struct {
	Name  string
	Email string
	Phone string `valid:"required"`
}

// VerifyPhoneNumber Verifies the user phone number
func VerifyPhoneNumber(w http.ResponseWriter, r *http.Request) {
	verifyPhoneNumber(w, r, common.ClientTypeIos)
}

// verifyPhoneNumber Verifies the user phone number
func verifyPhoneNumber(w http.ResponseWriter, r *http.Request, clientType int) {
	var verifyReq verifyPhoneNumberReq

	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &verifyReq)

	valid, err := govalidator.ValidateStruct(verifyReq)
	if valid != true {
		log.Printf("Invalid request params for verifying phone number")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	vCode, err := generateVerificationCode()
	if err != nil {
		m := fmt.Sprintf("Failed to generate OTP :%v", verifyReq.Phone)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var message strings.Builder
	message.Grow(32)
	if clientType == common.ClientTypeAndroid {
		fmt.Fprintf(&message, "[#]Thank you for signing up for iGrant.io! Your code is %s \n U1vUn/jAcoT", vCode)
	} else {
		fmt.Fprintf(&message, "Thank you for signing up for iGrant.io! Your code is %s", vCode)
	}

	err = sendPhoneVerificationMessage(verifyReq.Phone, verifyReq.Name, message.String())
	if err != nil {
		m := fmt.Sprintf("Failed to send sms to :%v", verifyReq.Phone)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	var o otp.Otp
	o.Name = verifyReq.Name
	o.Email = verifyReq.Email
	o.Phone = verifyReq.Phone
	o.Otp = vCode

	oldOtp, err := otp.SearchPhone(o.Phone)
	if err == nil {
		otp.Delete(oldOtp.ID.Hex())
	}

	o, err = otp.Add(o)
	if err != nil {
		m := fmt.Sprintf("Failed to store otp details")
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type verifyOtpReq struct {
	Phone string `valid:"required"`
	Otp   string `valid:"required"`
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
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	o, err := otp.SearchPhone(otpReq.Phone)
	if err != nil {
		valResp.Result = false
		valResp.Message = "Unregistered phone number: " + otpReq.Phone
		response, _ := json.Marshal(valResp)
		w.Header().Set("Content-Type", "application/json")
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
			common.HandleError(w, http.StatusInternalServerError, m, err)
			return
		}
	}

	response, _ := json.Marshal(valResp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
	return
}

type resetPasswordReq struct {
	Password string `valid:"required"`
}

type iamPasswordResetReq struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}

// ResetPassword Resets an user password
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	userName := token.GetUserName(r)
	userIamID := token.GetIamID(r)

	var resetReq resetPasswordReq
	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &resetReq)

	var status = http.StatusInternalServerError
	t, status, iamErr, err := getAdminToken()
	if err != nil {
		log.Printf("Failed to get admin token, user: %v registration", userName)
		handleError(w, userName, status, iamErr, err)
		return
	}

	valid, err := govalidator.ValidateStruct(resetReq)
	if !valid {
		log.Printf("Missing mandatory params required to reset password")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var e iamError
	iamReq := iamPasswordResetReq{"password", resetReq.Password, false}
	jsonReq, _ := json.Marshal(iamReq)
	req, err := http.NewRequest("PUT", iamConfig.URL+"/auth/admin/realms/"+iamConfig.Realm+"/users/"+userIamID+"/reset-password", bytes.NewBuffer(jsonReq))
	if err != nil {
		log.Printf("Failed to reset user:%v password ", userName)
		handleError(w, userName, status, iamErr, err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+t.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	//dump, err := httputil.DumpRequest(req, true)
	//dump, err := httputil.DumpRequestOut(req, true)
	//log.Printf("\n %q \n", dump)

	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to reset user:%v password ", userName)
		handleError(w, userName, status, iamErr, err)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(resp.Body)
		type errorMsg struct {
			ErrorMessage string `json:"errorMessage"`
		}
		var errMsg errorMsg
		json.Unmarshal(body, &errMsg)
		e.Error = errMsg.ErrorMessage
		e.ErrorType = "Reset password failed"
		response, _ := json.Marshal(e)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(response)
	}
	//TODO; json response needed for the creation successful.
	type resetPasswordResp struct {
		Msg string `json:"msg"`
	}

	response, _ := json.Marshal(resetPasswordResp{"User password resetted successfully"})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

type forgotPassword struct {
	Username string `valid:"required,email"`
}

// ForgotPassword User forgot the password, need to reset the password
func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var fp forgotPassword

	b, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(b, &fp)

	// validating request params
	valid, err := govalidator.ValidateStruct(fp)
	if !valid {
		log.Printf("Invalid request params for forgot password")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	log.Printf("User: %v forgot password", fp.Username)

	//Get user details from DB
	u, err := user.GetByEmail(fp.Username)
	if err != nil {
		log.Printf("User with %v doesnt exist", fp.Username)
		handleError(w, fp.Username, http.StatusNotFound, iamError{}, err)
		return
	}

	//curl  https://iam.igrant.io/auth/admin/realms/igrant-users/users/8b906c86-1ab2-4b32-becc-ba0349cb29ee/execute-actions-email -d '["UPDATE_PASSWORD"]' -X PUT -v
	var status = http.StatusInternalServerError
	t, status, iamErr, err := getAdminToken()
	if err != nil {
		log.Printf("Failed to get admin token, password forgot user: %v", fp.Username)
		handleError(w, fp.Username, status, iamErr, err)
		return
	}

	var e iamError
	//var iamReq = []byte(["UPDATE_PASSWORD"])
	var iamReq []string
	iamReq = append(iamReq, "UPDATE_PASSWORD")
	jsonReq, _ := json.Marshal(iamReq)

	req, err := http.NewRequest("PUT", iamConfig.URL+"/auth/admin/realms/"+iamConfig.Realm+"/users/"+u.IamID+"/execute-actions-email", bytes.NewBuffer(jsonReq))
	if err != nil {
		log.Printf("Failed to trigger forgot password action for user:%v", fp.Username)
		handleError(w, fp.Username, status, iamErr, err)
		return
	}

	req.Header.Add("Authorization", "Bearer "+t.AccessToken)
	req.Header.Add("Content-Type", "application/json")

	//dump, err := httputil.DumpRequest(req, true)
	//dump, err := httputil.DumpRequestOut(req, true)
	//log.Printf("\n %q \n", dump)

	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to trigger reset password email for user:%v", u.Name)
		handleError(w, fp.Username, status, iamErr, err)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		type errorMsg struct {
			ErrorMessage string `json:"errorMessage"`
		}
		var errMsg errorMsg
		json.Unmarshal(body, &errMsg)
		e.Error = errMsg.ErrorMessage
		e.ErrorType = "Forgot password handling failed"
		response, _ := json.Marshal(e)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(response)
		return
	}
	//TODO; json response needed for the creation successful.
	type resetPasswordResp struct {
		Msg string `json:"msg"`
	}

	response, _ := json.Marshal(resetPasswordResp{"User forgot password action handled successfully"})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

type logoutReq struct {
	RefreshToken string `valid:"required"`
	ClientID     string `valid:"required"`
}

// LogoutUser Logouts a user
func LogoutUser(w http.ResponseWriter, r *http.Request) {
	var lReq logoutReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &lReq)

	// validating request payload for logout
	valid, err := govalidator.ValidateStruct(lReq)

	if !valid {
		log.Printf("Failed to logout user:%v", token.GetUserName(r))
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	data := url.Values{}
	data.Set("refresh_token", lReq.RefreshToken)
	data.Add("client_id", lReq.ClientID)

	resp, err := http.PostForm(iamConfig.URL+"/auth/realms/"+iamConfig.Realm+"/protocol/openid-connect/logout", data)
	if err != nil {
		m := fmt.Sprintf("Failed to logout user:%v", token.GetUserName(r))
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		m := fmt.Sprintf("Failed to logout user:%v", token.GetUserName(r))
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	if resp.StatusCode != http.StatusNoContent {
		var e iamError
		json.Unmarshal(body, &e)
		response, _ := json.Marshal(e)
		w.WriteHeader(resp.StatusCode)
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
		return
	}
	u, err := user.Get(token.GetUserID(r))
	if err != nil {
		m := fmt.Sprintf("Failed to locate user:%v", token.GetUserName(r))
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	if len(u.Roles) != 0 {
		actionLog := fmt.Sprintf("%v logged out", u.Email)
		actionlog.LogOrgSecurityCalls(token.GetUserID(r), token.GetUserName(r), u.Roles[0].OrgID, actionLog)
	}
	w.WriteHeader(http.StatusNoContent)
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

func sendPhoneVerificationMessage(msgTo string, name string, message string) error {
	urlStr := "https://api.twilio.com/2010-04-01/Accounts/" + twilioConfig.AccountSid + "/Messages.json"

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
	req.SetBasicAuth(twilioConfig.AccountSid, twilioConfig.AuthToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

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
