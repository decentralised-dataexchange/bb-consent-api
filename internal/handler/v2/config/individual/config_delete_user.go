package individual

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/iam"
	"github.com/bb-consent/api/internal/individual"
	"github.com/gorilla/mux"
)

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

	// Unregister individual in iam
	err = iam.UnregisterIndividual(individual.IamId)
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
