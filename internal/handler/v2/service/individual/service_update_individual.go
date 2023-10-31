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

func updateIndividualFromUpdateIndividualServiceRequestBody(requestBody updateServiceIndividualReq, tobeUpdatedIndividual individual.Individual) individual.Individual {
	tobeUpdatedIndividual.ExternalId = requestBody.Individual.ExternalId
	tobeUpdatedIndividual.ExternalIdType = requestBody.Individual.ExternalIdType
	tobeUpdatedIndividual.IdentityProviderId = requestBody.Individual.IdentityProviderId
	tobeUpdatedIndividual.Name = requestBody.Individual.Name
	tobeUpdatedIndividual.Email = requestBody.Individual.Email
	tobeUpdatedIndividual.Phone = requestBody.Individual.Phone

	return tobeUpdatedIndividual
}

type updateServiceIndividualReq struct {
	Individual individual.Individual `json:"individual"`
}

type updateServiceIndividualResp struct {
	Individual individual.Individual `json:"individual"`
}

func ServiceUpdateIndividual(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	individualId := mux.Vars(r)[config.IndividualId]
	individualId = common.Sanitize(individualId)

	// Request body
	var individualReq updateServiceIndividualReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &individualReq)

	// Validate request body
	// validating request payload
	valid, err := govalidator.ValidateStruct(individualReq)
	if !valid {
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

	err = iam.UpdateIamIndividual(individualReq.Individual.Name, tobeUpdatedIndividual.IamId, individualReq.Individual.Email)
	if err != nil {
		m := fmt.Sprintf("Failed to update IAM user by id:%v", individualId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	tobeUpdatedIndividual = updateIndividualFromUpdateIndividualServiceRequestBody(individualReq, tobeUpdatedIndividual)

	savedIndividual, err := individualRepo.Update(tobeUpdatedIndividual)
	if err != nil {
		m := fmt.Sprintf("Failed to update individual: %v", individualId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := updateServiceIndividualResp{
		Individual: savedIndividual,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
