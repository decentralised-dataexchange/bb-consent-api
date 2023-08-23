package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/actionlog"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/email"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/user"
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
