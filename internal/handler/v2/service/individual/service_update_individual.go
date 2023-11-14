package individual

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/iam"
	"github.com/bb-consent/api/internal/individual"
	"github.com/gorilla/mux"
)

func updateIndividualFromUpdateIndividualServiceRequestBody(requestBody updateServiceIndividualReq, tobeUpdatedIndividual individual.Individual) individual.Individual {

	if len(strings.TrimSpace(requestBody.Individual.ExternalId)) > 1 {
		tobeUpdatedIndividual.ExternalId = requestBody.Individual.ExternalId
	}
	if len(strings.TrimSpace(requestBody.Individual.ExternalIdType)) > 1 {
		tobeUpdatedIndividual.ExternalIdType = requestBody.Individual.ExternalIdType
	}
	if len(strings.TrimSpace(requestBody.Individual.IdentityProviderId)) > 1 {
		tobeUpdatedIndividual.IdentityProviderId = requestBody.Individual.IdentityProviderId
	}
	if len(strings.TrimSpace(requestBody.Individual.Name)) > 1 {
		tobeUpdatedIndividual.Name = requestBody.Individual.Name
	}
	if len(strings.TrimSpace(requestBody.Individual.Email)) > 1 {
		tobeUpdatedIndividual.Email = requestBody.Individual.Email
	}
	if len(strings.TrimSpace(requestBody.Individual.Phone)) > 1 {
		tobeUpdatedIndividual.Phone = requestBody.Individual.Phone
	}

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

	currentIndividual, err := individualRepo.Get(individualId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch individual: %v", individualId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	toBeUpdatedIndividual := updateIndividualFromUpdateIndividualServiceRequestBody(individualReq, currentIndividual)

	if currentIndividual.Name != toBeUpdatedIndividual.Name || currentIndividual.Email != toBeUpdatedIndividual.Email {

		if len(strings.TrimSpace(currentIndividual.IamId)) > 1 {
			// Update individual in iam
			err = iam.UpdateIamIndividual(toBeUpdatedIndividual.Name, currentIndividual.IamId, toBeUpdatedIndividual.Email)
			if err != nil {
				m := fmt.Sprintf("Failed to update IAM user by id:%v", individualId)
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}
		} else if len(strings.TrimSpace(toBeUpdatedIndividual.Email)) > 1 {
			iamId, err := iam.RegisterUser(toBeUpdatedIndividual.Email, toBeUpdatedIndividual.Name)
			if err != nil {
				m := fmt.Sprintf("Failed to create IAM user by id:%v", individualId)
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}
			toBeUpdatedIndividual.IamId = iamId
		}

	}

	savedIndividual, err := individualRepo.Update(toBeUpdatedIndividual)
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
