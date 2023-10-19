package idp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/v2/idp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type addIdpReq struct {
	Idp idp.IdentityProvider `json:"idp" valid:"required"`
}

type addIdpResp struct {
	Idp idp.IdentityProvider `json:"idp" valid:"required"`
}

func ConfigCreateIdp(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Request body
	var idpReq addIdpReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &idpReq)

	// Repository
	idpRepo := idp.IdentityProviderRepository{}
	idpRepo.Init(organisationId)

	exists, err := idpRepo.IsIdentityProviderExist()
	if err != nil || exists >= 1 {
		m := fmt.Sprintf("External IDP exists; Try to update instead of create; Failed to create identity provider for %v", organisationId)
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	// validating request payload
	valid, err := govalidator.ValidateStruct(idpReq)
	if !valid {
		m := fmt.Sprintf("Missing mandatory params for creating identity provider for org:%v\n", organisationId)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	var newIdentityProvider idp.IdentityProvider

	// OpenID config
	newIdentityProvider.Id = primitive.NewObjectID().Hex()
	newIdentityProvider.AuthorizationURL = idpReq.Idp.AuthorizationURL
	newIdentityProvider.TokenURL = idpReq.Idp.TokenURL
	newIdentityProvider.LogoutURL = idpReq.Idp.LogoutURL
	newIdentityProvider.ClientID = idpReq.Idp.ClientID
	newIdentityProvider.ClientSecret = idpReq.Idp.ClientSecret
	newIdentityProvider.JWKSURL = idpReq.Idp.JWKSURL
	newIdentityProvider.UserInfoURL = idpReq.Idp.UserInfoURL
	newIdentityProvider.DefaultScope = idpReq.Idp.DefaultScope
	newIdentityProvider.IssuerUrl = idpReq.Idp.IssuerUrl
	newIdentityProvider.OrganisationId = organisationId
	newIdentityProvider.IsDeleted = false

	savedIdp, err := idpRepo.Add(newIdentityProvider)
	if err != nil {
		m := fmt.Sprintf("Failed to create new idp: %v", organisationId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := addIdpResp{
		Idp: savedIdp,
	}
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
