package onboard

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/iam"
	"github.com/bb-consent/api/internal/idp"
	"github.com/bb-consent/api/internal/org"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

func refreshTokenForExternalIdpIssuedToken(refreshToken string) (*oauth2.Token, error) {
	var token *oauth2.Token
	// Get organisation
	organisation, err := org.GetFirstOrganization()
	if err != nil {
		return token, err
	}

	// Repository
	idpRepo := idp.IdentityProviderRepository{}
	idpRepo.Init(organisation.ID.Hex())

	// Fetch IDP details based on org Id
	idp, err := idpRepo.GetByOrgId()
	if err != nil {
		return token, err
	}

	provider, err := oidc.NewProvider(context.Background(), idp.IssuerUrl)
	if err != nil {
		return token, err
	}

	// Initialize the OAuth2 configuration
	oauth2Config := &oauth2.Config{
		ClientID:     idp.ClientID,
		ClientSecret: idp.ClientSecret,
		Endpoint:     provider.Endpoint(),
	}

	ts := oauth2Config.TokenSource(context.Background(), &oauth2.Token{RefreshToken: refreshToken})
	tok, err := ts.Token()
	return tok, err
}

type tokenReq struct {
	RefreshToken string `valid:"required" json:"refreshToken"`
	ClientID     string `valid:"required" json:"clientId"`
}

type refreshTokenResp struct {
	AccessToken      string `json:"accessToken"`
	ExpiresIn        int    `json:"expiresIn"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
	RefreshToken     string `json:"refreshToken"`
	TokenType        string `json:"tokenType"`
}

// OnboardRefreshToken
func OnboardRefreshToken(w http.ResponseWriter, r *http.Request) {
	var tReq tokenReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &tReq)

	// validating request payload for refreshing tokens
	valid, err := govalidator.ValidateStruct(tReq)

	if !valid {
		log.Printf("Failed to refresh token")
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var resp refreshTokenResp
	if tReq.ClientID == iam.IamConfig.ClientId {
		// Refresh token for consent bb users
		client := iam.GetClient()

		t, err := iam.RefreshToken(tReq.ClientID, tReq.RefreshToken, iam.IamConfig.Realm, client)
		if err != nil {
			log.Printf("Failed to refresh token")
			common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
			return
		}
		resp = refreshTokenResp{
			AccessToken:      t.AccessToken,
			ExpiresIn:        t.ExpiresIn,
			RefreshExpiresIn: t.RefreshExpiresIn,
			RefreshToken:     t.RefreshToken,
			TokenType:        t.TokenType,
		}

	} else {
		// Refresh Token For External Idp issued Token
		t, err := refreshTokenForExternalIdpIssuedToken(tReq.RefreshToken)
		if err != nil {
			log.Printf("Failed to refresh token for external idp")
			common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
			return
		}
		resp = refreshTokenResp{
			AccessToken:      t.AccessToken,
			ExpiresIn:        t.Expiry.Second(),
			RefreshExpiresIn: t.Expiry.Second(),
			RefreshToken:     t.RefreshToken,
			TokenType:        t.TokenType,
		}
	}

	response, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}
