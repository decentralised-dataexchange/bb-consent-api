package individual

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Nerzal/gocloak/v13"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/iam"
	"github.com/bb-consent/api/src/v2/individual"
	"github.com/gorilla/mux"
)

// unregisterUser Unregisters an existing user
func unregisterUser(iamUserID string, adminToken string, client *gocloak.GoCloak) error {
	err := client.DeleteUser(context.Background(), adminToken, iam.IamConfig.Realm, iamUserID)
	return err
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

	client := getClient()

	t, err := getAdminToken(client)
	if err != nil {
		log.Printf("Failed to get admin token, user: %v registration", individualId)
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	err = unregisterUser(individual.IamId, t.AccessToken, client)
	if err != nil {
		m := fmt.Sprintf("Failed to unregister user: %v err: %v", individualId, err)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
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
