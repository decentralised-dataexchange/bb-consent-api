package idp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/idp"
	"github.com/gorilla/mux"
)

type deleteIdpResp struct {
	Idp idp.IdentityProvider `json:"idp" valid:"required"`
}

// DeleteIdentityProvider Delete external identity provider for an organisation
func DeleteIdentityProvider(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Path params
	idpId := mux.Vars(r)[config.IdpId]
	idpId = common.Sanitize(idpId)

	// Repository
	idpRepo := idp.IdentityProviderRepository{}
	idpRepo.Init(organisationId)

	toBeDeletedIdp, err := idpRepo.Get(idpId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch identity provider: %v", idpId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	toBeDeletedIdp.IsDeleted = true

	savedIdp, err := idpRepo.Update(toBeDeletedIdp)
	if err != nil {
		m := fmt.Sprintf("Failed to delete idp: %v", idpId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := deleteIdpResp{
		Idp: savedIdp,
	}
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
