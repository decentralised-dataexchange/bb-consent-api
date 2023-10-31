package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/idp"
	"github.com/gorilla/mux"
)

type updateIdpReq struct {
	Idp idp.IdentityProvider `json:"idp" valid:"required"`
}

type updateIdpResp struct {
	Idp idp.IdentityProvider `json:"idp" valid:"required"`
}

// UpdateIdentityProvider Update external identity provider for an organisation
func UpdateIdentityProvider(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Path params
	idpId := mux.Vars(r)[config.IdpId]
	idpId = common.Sanitize(idpId)

	// Request body
	var idpReq updateIdpReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &idpReq)

	// validating request payload
	valid, err := govalidator.ValidateStruct(idpReq)
	if !valid {
		m := fmt.Sprintf("Missing mandatory params for creating identity provider for org:%v\n", organisationId)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Repository
	idpRepo := idp.IdentityProviderRepository{}
	idpRepo.Init(organisationId)

	toBeUpdatedIndentityProvider, err := idpRepo.Get(idpId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch identity provider: %v", idpId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// OpenID config
	toBeUpdatedIndentityProvider.AuthorizationURL = idpReq.Idp.AuthorizationURL
	toBeUpdatedIndentityProvider.TokenURL = idpReq.Idp.TokenURL
	toBeUpdatedIndentityProvider.LogoutURL = idpReq.Idp.LogoutURL
	toBeUpdatedIndentityProvider.ClientID = idpReq.Idp.ClientID
	toBeUpdatedIndentityProvider.ClientSecret = idpReq.Idp.ClientSecret
	toBeUpdatedIndentityProvider.JWKSURL = idpReq.Idp.JWKSURL
	toBeUpdatedIndentityProvider.UserInfoURL = idpReq.Idp.UserInfoURL
	toBeUpdatedIndentityProvider.DefaultScope = idpReq.Idp.DefaultScope
	toBeUpdatedIndentityProvider.IssuerUrl = idpReq.Idp.IssuerUrl

	savedIdp, err := idpRepo.Update(toBeUpdatedIndentityProvider)
	if err != nil {
		m := fmt.Sprintf("Failed to update idp: %v", organisationId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := updateIdpResp{
		Idp: savedIdp,
	}
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
