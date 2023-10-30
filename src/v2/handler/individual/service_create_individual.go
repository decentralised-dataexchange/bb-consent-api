package individual

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/individual"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func createIamRegisterRequestFromAddRequestBody(requestBody addServiceIndividualReq, iamRegReq iamIndividualRegisterReq) iamIndividualRegisterReq {

	iamRegReq.Username = requestBody.Individual.Email
	iamRegReq.Firstname = requestBody.Individual.Name
	iamRegReq.Email = requestBody.Individual.Email
	iamRegReq.Enabled = true
	iamRegReq.RequiredActions = []string{"UPDATE_PASSWORD"}

	return iamRegReq
}
func updateIndividualFromAddRequestBody(requestBody addServiceIndividualReq, newIndividual individual.Individual) individual.Individual {
	newIndividual.ExternalId = requestBody.Individual.ExternalId
	newIndividual.ExternalIdType = requestBody.Individual.ExternalIdType
	newIndividual.IdentityProviderId = requestBody.Individual.IdentityProviderId
	newIndividual.Name = requestBody.Individual.Name
	newIndividual.Email = requestBody.Individual.Email
	newIndividual.Phone = requestBody.Individual.Phone

	return newIndividual
}

type addServiceIndividualReq struct {
	Individual individual.Individual `json:"individual"`
}

type addServiceIndividualResp struct {
	Individual individual.Individual `json:"individual"`
}

// ServiceCreateIndividual
func ServiceCreateIndividual(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Request body
	var individualReq addServiceIndividualReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &individualReq)

	// Validate request body
	valid, err := govalidator.ValidateStruct(individualReq)
	if !valid {
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var iamRegReq iamIndividualRegisterReq

	iamRegReq = createIamRegisterRequestFromAddRequestBody(individualReq, iamRegReq)

	client := getClient()

	t, err := getAdminToken(client)
	if err != nil {
		log.Printf("Failed to get admin token, user: %v registration", individualReq.Individual.Email)
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	iamId, err := registerUser(iamRegReq, t.AccessToken, client)
	if err != nil {
		log.Printf("Failed to register user: %v err: %v", individualReq.Individual.Email, err)
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var newIndividual individual.Individual
	newIndividual.Id = primitive.NewObjectID()
	newIndividual.IamId = iamId
	newIndividual = updateIndividualFromAddRequestBody(individualReq, newIndividual)
	newIndividual.OrganisationId = organisationId
	newIndividual.IsDeleted = false
	newIndividual.IsOnboardedFromId = false

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	// Save the individual to db
	savedIndividual, err := individualRepo.Add(newIndividual)
	if err != nil {
		m := fmt.Sprintf("Failed to create new individual: %v", newIndividual.Name)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := addServiceIndividualResp{
		Individual: savedIndividual,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
