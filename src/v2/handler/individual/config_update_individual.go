package individual

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/individual"
	"github.com/gorilla/mux"
)

func validateUpdateIndividualRequestBody(IndividualReq updateIndividualReq) error {
	// validating request payload
	valid, err := govalidator.ValidateStruct(IndividualReq)
	if !valid {
		return err
	}

	return nil
}

type iamIndividualUpdateReq struct {
	Username  string `json:"username"`
	Firstname string `json:"firstName"`
	Email     string `json:"email"`
}

// updateIamIndividual Update user info on IAM server end.
func updateIamIndividual(iamUpdateReq iamIndividualUpdateReq, iamID string) error {

	t, _, _, err := getAdminToken()
	if err != nil {
		log.Printf("Failed to get admin token, user: %v update err:%v", iamUpdateReq.Firstname, err)
		return err
	}

	jsonReq, _ := json.Marshal(iamUpdateReq)
	req, err := http.NewRequest("PUT", iamConfig.URL+"/admin/realms/"+iamConfig.Realm+"/users/"+iamID, bytes.NewBuffer(jsonReq))
	if err != nil {
		log.Printf("Failed to form update user request err:%v", err)
		return err
	}

	req.Header.Add("Authorization", "Bearer "+t.AccessToken)
	req.Header.Add(config.ContentTypeHeader, config.ContentTypeJSON)

	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		log.Printf("Failed to update user with status code: %v", resp.StatusCode)
		return errors.New("failed to update user")
	}
	return nil
}

func updateIamUpdateRequestFromRequestBody(requestBody updateIndividualReq) iamIndividualUpdateReq {
	var iamIndividualReq iamIndividualUpdateReq

	iamIndividualReq.Username = requestBody.Individual.Email
	iamIndividualReq.Firstname = requestBody.Individual.Name
	iamIndividualReq.Email = requestBody.Individual.Email

	return iamIndividualReq
}

func updateIndividualFromUpdateIndividualRequestBody(requestBody updateIndividualReq, tobeUpdatedIndividual individual.Individual) individual.Individual {
	tobeUpdatedIndividual.ExternalId = requestBody.Individual.ExternalId
	tobeUpdatedIndividual.ExternalIdType = requestBody.Individual.ExternalIdType
	tobeUpdatedIndividual.IdentityProviderId = requestBody.Individual.IdentityProviderId
	tobeUpdatedIndividual.Name = requestBody.Individual.Name
	tobeUpdatedIndividual.Email = requestBody.Individual.Email
	tobeUpdatedIndividual.Phone = requestBody.Individual.Phone

	return tobeUpdatedIndividual
}

type updateIndividualReq struct {
	Individual individual.Individual `json:"individual"`
}

type updateIndividualResp struct {
	Individual individual.Individual `json:"individual"`
}

func ConfigUpdateIndividual(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	individualId := mux.Vars(r)[config.IndividualId]
	individualId = common.Sanitize(individualId)

	// Request body
	var individualReq updateIndividualReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &individualReq)

	// Validate request body
	err := validateUpdateIndividualRequestBody(individualReq)
	if err != nil {
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	tobeUpdatedIndividual, err := individualRepo.Get(individualId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch individual: %v", individualId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	iamUpdateReq := updateIamUpdateRequestFromRequestBody(individualReq)

	err = updateIamIndividual(iamUpdateReq, tobeUpdatedIndividual.IamId)
	if err != nil {
		m := fmt.Sprintf("Failed to update IAM user by id:%v", individualId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	tobeUpdatedIndividual = updateIndividualFromUpdateIndividualRequestBody(individualReq, tobeUpdatedIndividual)

	savedIndividual, err := individualRepo.Update(tobeUpdatedIndividual)
	if err != nil {
		m := fmt.Sprintf("Failed to update individual: %v", individualId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := updateIndividualResp{
		Individual: savedIndividual,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
