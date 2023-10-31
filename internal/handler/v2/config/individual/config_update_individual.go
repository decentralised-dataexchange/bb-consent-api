package individual

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/iam"
	"github.com/bb-consent/api/internal/individual"
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

	// Update individual in iam
	err = iam.UpdateIamIndividual(individualReq.Individual.Name, tobeUpdatedIndividual.IamId, individualReq.Individual.Email)
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
