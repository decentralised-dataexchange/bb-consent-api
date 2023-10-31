package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/idp"
	"github.com/gorilla/mux"
)

type serviceIdp struct {
	Id        string `json:"id"`
	LogoutUrl string `json:"logoutUrl"`
	ClientId  string `json:"clientId"`
}

type readIdpResp struct {
	Idp serviceIdp `json:"idp"`
}

// ServiceReadIdp
func ServiceReadIdp(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Path params
	idpId := mux.Vars(r)[config.IdpId]
	idpId = common.Sanitize(idpId)

	// Repository
	idpRepo := idp.IdentityProviderRepository{}
	idpRepo.Init(organisationId)

	idp, err := idpRepo.Get(idpId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch identity provider: %v", idpId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	idpResp := serviceIdp{
		Id:        idp.Id.Hex(),
		LogoutUrl: idp.LogoutURL,
		ClientId:  idp.ClientID,
	}

	resp := readIdpResp{
		Idp: idpResp,
	}
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
