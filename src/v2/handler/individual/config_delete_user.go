package individual

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/individual"
	"github.com/gorilla/mux"
)

// unregisterUser Unregisters an existing user
func unregisterUser(iamUserID string, adminToken string) (int, iamError, error) {
	var e iamError
	var status = http.StatusInternalServerError
	req, err := http.NewRequest("DELETE", iamConfig.URL+"/admin/realms/"+iamConfig.Realm+"/users/"+iamUserID, nil)
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
	if resp.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		type errorMsg struct {
			ErrorMessage string `json:"errorMessage"`
		}
		var errMsg errorMsg
		json.Unmarshal(body, &errMsg)
		e.Error = errMsg.ErrorMessage
		e.ErrorType = "Creation failed"
		return resp.StatusCode, e, errors.New("failed to unregister user")
	}
	return resp.StatusCode, e, err
}

type deleteIndividualResp struct {
	Individual individual.Individual `json:"individual"`
}

func ConfigDeleteIndividual(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	individualId := mux.Vars(r)[config.IndividualId]
	individualId = common.Sanitize(individualId)

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	individual, err := individualRepo.Get(individualId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch individual: %v", individualId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	t, status, iamErr, err := getAdminToken()
	if err != nil {
		log.Printf("Failed to get admin token, user: %v registration", individualId)
		handleError(w, individualId, status, iamErr, err)
		return
	}

	status, iamErr, err = unregisterUser(individual.IamId, t.AccessToken)
	if err != nil {
		log.Printf("Failed to unregister user: %v err: %v", individualId, err)
		handleError(w, individualId, status, iamErr, err)
		return
	}

	individual.IsDeleted = true

	savedIndividual, err := individualRepo.Update(individual)
	if err != nil {
		m := fmt.Sprintf("Failed to update individual: %v", individualId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := deleteIndividualResp{
		Individual: savedIndividual,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
