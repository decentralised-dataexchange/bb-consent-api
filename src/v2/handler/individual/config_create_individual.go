package individual

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/iam"
	"github.com/bb-consent/api/src/v2/individual"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getClient() *gocloak.GoCloak {
	client := gocloak.NewClient(iam.IamConfig.URL)
	return client
}

func getToken(username string, password string, realm string, client *gocloak.GoCloak) (*gocloak.JWT, error) {
	ctx := context.Background()
	token, err := client.LoginAdmin(ctx, username, password, realm)
	if err != nil {
		return token, err
	}

	return token, err
}

func getAdminToken(client *gocloak.GoCloak) (*gocloak.JWT, error) {
	t, err := getToken(iam.IamConfig.AdminUser, iam.IamConfig.AdminPassword, "master", client)
	return t, err
}

// registerUser Registers a new user
func registerUser(iamRegReq iamIndividualRegisterReq, adminToken string, client *gocloak.GoCloak) (string, error) {
	user := gocloak.User{
		FirstName: &iamRegReq.Firstname,
		Email:     &iamRegReq.Email,
		Enabled:   gocloak.BoolP(true),
		Username:  &iamRegReq.Email,
	}

	iamId, err := client.CreateUser(context.Background(), adminToken, iam.IamConfig.Realm, user)
	if err != nil {
		return "", err
	}
	return iamId, err
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
	newIndividual = updateIndividualFromRequestBody(individualReq, newIndividual)
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

	resp := addIndividualResp{
		Individual: savedIndividual,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
