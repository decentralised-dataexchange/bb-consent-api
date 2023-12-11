package service

import (
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/idp"
)

type serviceIdp struct {
	Id               string `json:"id" bson:"_id,omitempty"`
	IssuerUrl        string `json:"issuerUrl"`
	AuthorizationURL string `json:"authorisationUrl"`
	TokenURL         string `json:"tokenUrl"`
	LogoutURL        string `json:"logoutUrl"`
	ClientID         string `json:"clientId"`
	JWKSURL          string `json:"jwksUrl"`
	UserInfoURL      string `json:"userInfoUrl"`
	DefaultScope     string `json:"defaultScope"`
}

type readIdpResp struct {
	Idp interface{} `json:"idp"`
}

// ServiceReadIdp
func ServiceReadIdp(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Repository
	idpRepo := idp.IdentityProviderRepository{}
	idpRepo.Init(organisationId)

	idp, err := idpRepo.GetByOrgId()
	if err != nil {
		resp := readIdpResp{
			Idp: nil,
		}
		common.ReturnHTTPResponse(resp, w)
		return
	}
	idpResp := serviceIdp{
		Id:               idp.Id,
		IssuerUrl:        idp.IssuerUrl,
		AuthorizationURL: idp.AuthorizationURL,
		TokenURL:         idp.TokenURL,
		LogoutURL:        idp.LogoutURL,
		ClientID:         idp.ClientID,
		JWKSURL:          idp.JWKSURL,
		UserInfoURL:      idp.UserInfoURL,
		DefaultScope:     idp.DefaultScope,
	}

	resp := readIdpResp{
		Idp: idpResp,
	}
	common.ReturnHTTPResponse(resp, w)
}
