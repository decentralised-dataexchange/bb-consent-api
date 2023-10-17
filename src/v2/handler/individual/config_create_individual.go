package individual

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/individual"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var timeout time.Duration

var iamConfig config.Iam

// IamInit Initialize the IAM handler
func IamInit(config *config.Configuration) {
	iamConfig = config.Iam
	timeout = time.Duration(time.Duration(iamConfig.Timeout) * time.Second)

}

type iamToken struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
}

type iamError struct {
	ErrorType string `json:"errorCode"`
	Error     string `json:"errorDescription"`
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

	resp, err := http.PostForm(iamConfig.URL+"/realms/"+realm+"/protocol/openid-connect/token", data)
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

func getAdminToken() (iamToken, int, iamError, error) {
	t, status, iamErr, err := getToken(iamConfig.AdminUser, iamConfig.AdminPassword, "admin-cli", "master")
	return t, status, iamErr, err
}

// registerUser Registers a new user
func registerUser(iamRegReq iamIndividualRegisterReq, adminToken string) (int, iamError, error) {
	var e iamError
	var status = http.StatusInternalServerError
	jsonReq, _ := json.Marshal(iamRegReq)
	req, err := http.NewRequest("POST", iamConfig.URL+"/admin/realms/"+iamConfig.Realm+"/users", bytes.NewBuffer(jsonReq))
	if err != nil {
		return status, e, err
	}

	req.Header.Add("Authorization", "Bearer "+adminToken)
	req.Header.Add(config.ContentTypeHeader, config.ContentTypeJSON)

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
	req, err := http.NewRequest("GET", iamConfig.URL+"/admin/realms/"+iamConfig.Realm+"/users"+"?username="+userName, nil)
	if err != nil {
		return userID, status, e, err
	}
	//log.Printf("token: %v", t)
	req.Header.Add("Authorization", "Bearer "+adminToken)
	req.Header.Add(config.ContentTypeHeader, config.ContentTypeJSON)

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

func createIamRegisterRequestFromRequestBody(requestBody addIndividualReq, iamRegReq iamIndividualRegisterReq) iamIndividualRegisterReq {

	iamRegReq.Username = requestBody.Individual.Email
	iamRegReq.Firstname = requestBody.Individual.Name
	iamRegReq.Email = requestBody.Individual.Email
	iamRegReq.Enabled = true
	iamRegReq.RequiredActions = []string{"UPDATE_PASSWORD"}

	return iamRegReq
}
func updateIndividualFromRequestBody(requestBody addIndividualReq, newIndividual individual.Individual) individual.Individual {
	newIndividual.ExternalId = requestBody.Individual.ExternalId
	newIndividual.ExternalIdType = requestBody.Individual.ExternalIdType
	newIndividual.IdentityProviderId = requestBody.Individual.IdentityProviderId
	newIndividual.Name = requestBody.Individual.Name
	newIndividual.Email = requestBody.Individual.Email
	newIndividual.Phone = requestBody.Individual.Phone

	return newIndividual
}
func handleError(w http.ResponseWriter, user string, status int, iamErr iamError, err error) {
	if (iamError{}) != iamErr {
		log.Printf("Failed to register err:%v", err)
		resp, _ := json.Marshal(iamErr)
		w.WriteHeader(status)
		w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
		w.Write(resp)
		return
	}
	m := fmt.Sprintf("Failed to create user:%v", user)
	common.HandleErrorV2(w, status, m, err)
}

type iamIndividualRegisterReq struct {
	Username        string   `json:"username"`
	Firstname       string   `json:"firstName"`
	Email           string   `json:"email"`
	Enabled         bool     `json:"enabled"`
	RequiredActions []string `json:"requiredActions"`
}

func validateAddIndividualRequestBody(IndividualReq addIndividualReq) error {
	// validating request payload
	valid, err := govalidator.ValidateStruct(IndividualReq)
	if !valid {
		return err
	}

	return nil
}

type addIndividualReq struct {
	Individual individual.Individual `json:"individual"`
}

type addIndividualResp struct {
	Individual individual.Individual `json:"individual"`
}

// ConfigCreateIndividual
func ConfigCreateIndividual(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Request body
	var individualReq addIndividualReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &individualReq)

	// Validate request body
	err := validateAddIndividualRequestBody(individualReq)
	if err != nil {
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var iamRegReq iamIndividualRegisterReq

	iamRegReq = createIamRegisterRequestFromRequestBody(individualReq, iamRegReq)

	t, status, iamErr, err := getAdminToken()
	if err != nil {
		log.Printf("Failed to get admin token, user: %v registration", individualReq.Individual.Email)
		handleError(w, individualReq.Individual.Email, status, iamErr, err)
		return
	}

	status, iamErr, err = registerUser(iamRegReq, t.AccessToken)
	if err != nil {
		log.Printf("Failed to register user: %v err: %v", individualReq.Individual.Email, err)
		handleError(w, individualReq.Individual.Email, status, iamErr, err)
		return
	}

	userIamID, status, iamErr, err := getUserIamID(individualReq.Individual.Email, t.AccessToken)
	if err != nil {
		log.Printf("Failed to get userID for user: %v err: %v", individualReq.Individual.Email, err)
		handleError(w, individualReq.Individual.Email, status, iamErr, err)
		return
	}

	var newIndividual individual.Individual
	newIndividual.Id = primitive.NewObjectID().Hex()
	newIndividual.IamId = userIamID
	newIndividual = updateIndividualFromRequestBody(individualReq, newIndividual)
	newIndividual.OrganisationId = organisationId
	newIndividual.IsDeleted = false

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	// Save the data attribute to db
	savedIndividual, err := individualRepo.Add(newIndividual)
	if err != nil {
		m := fmt.Sprintf("Failed to create new individual: %v", newIndividual.Name)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := addIndividualResp{
		Individual: savedIndividual,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
