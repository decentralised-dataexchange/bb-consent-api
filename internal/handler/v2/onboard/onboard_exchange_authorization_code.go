package onboard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/idp"
	"github.com/bb-consent/api/internal/individual"
	"github.com/coreos/go-oidc/v3/oidc"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
)

type userInfoResp struct {
	Subject       string `json:"subject"`
	Profile       string `json:"profile"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
}
type tResp struct {
	AccessToken      string `json:"accessToken"`
	ExpiresIn        int    `json:"expiresIn"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
	RefreshToken     string `json:"refreshToken"`
	TokenType        string `json:"tokenType"`
}

type exchangeAuthorizationResp struct {
	UserInfo userInfoResp `json:"userInfo"`
	Token    tResp        `json:"token"`
}

type exchangeAuthorizationReq struct {
	RedirectUri string `json:"redirectUri" valid:"required"`
}

// ExchangeAuthorizationCode Exchange the authorization code for an access token
func ExchangeAuthorizationCode(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	var req exchangeAuthorizationReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &req)

	// validating the request payload
	valid, err := govalidator.ValidateStruct(req)

	if !valid {
		log.Printf("Invalid request params!")
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Query params
	oauthAuthorizationCode := r.URL.Query().Get(config.AuthorisationCode)

	if len(strings.TrimSpace(oauthAuthorizationCode)) == 0 {
		log.Printf("Missing mandatory query params for exchanging authorization code \n")
		m := fmt.Sprintf("Failed to exchange authorization code for org:%v", organisationId)
		common.HandleError(w, http.StatusNotFound, m, errors.New(m))
		return
	}

	// Repository
	idpRepo := idp.IdentityProviderRepository{}
	idpRepo.Init(organisationId)

	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	// Fetch IDP details based on org Id
	idp, err := idpRepo.GetByOrgId()
	if err != nil {
		m := fmt.Sprintf("failed to fetch idp of individual:%v", organisationId)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	provider, err := oidc.NewProvider(context.Background(), idp.IssuerUrl)
	if err != nil {
		m := "failed to initialize oidc provider"
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	// Initialize the OAuth2 configuration
	oauth2Config := &oauth2.Config{
		ClientID:     idp.ClientID,
		ClientSecret: idp.ClientSecret,
		RedirectURL:  req.RedirectUri,
		Endpoint:     provider.Endpoint(),
	}

	// Exchange authorisation code for access token from organisation's IDP
	token, err := oauth2Config.Exchange(context.Background(), oauthAuthorizationCode)
	if err != nil {
		m := "Invalid grant code not valid"
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}
	var individualEmail string
	var individualExternalId string
	var individualEmailVerified bool
	var individualProfile string

	//  Fetch user information from the UserInfo endpoint
	userInfo, err := provider.UserInfo(context.Background(), oauth2.StaticTokenSource(token))
	if err != nil {

		jwks := oidc.NewRemoteKeySet(context.Background(), idp.JWKSURL)
		c := oidc.NewVerifier(idp.IssuerUrl, jwks, &oidc.Config{SkipClientIDCheck: true})
		tokenPayload, err := c.Verify(context.Background(), token.AccessToken)
		if err != nil {
			m := "Failed to fetch user information from token"
			common.HandleError(w, http.StatusNotFound, m, err)
			return
		}
		var claims struct {
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
		}
		if err := tokenPayload.Claims(&claims); err != nil {
			m := "Failed to fetch user information from token"
			common.HandleError(w, http.StatusNotFound, m, err)
			return
		}
		individualEmail = claims.Email
		individualExternalId = tokenPayload.Subject
		individualEmailVerified = claims.EmailVerified

	} else {
		individualEmail = userInfo.Email
		individualExternalId = userInfo.Subject
		individualEmailVerified = userInfo.EmailVerified
		individualProfile = userInfo.Profile
	}

	_, err = individualRepo.GetByExternalId(individualExternalId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Create individual if not present
			createIndividualFromIdp(individualEmail, individualExternalId, organisationId, idp.Id.Hex())
		} else {
			m := fmt.Sprintf("Failed to fetch individual: %v", individualExternalId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
	}
	t := tResp{
		AccessToken:  token.AccessToken,
		ExpiresIn:    token.Expiry.Minute(),
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
	}
	u := userInfoResp{
		Subject:       individualExternalId,
		Profile:       individualProfile,
		Email:         individualEmail,
		EmailVerified: individualEmailVerified,
	}

	response, _ := json.Marshal(exchangeAuthorizationResp{u, t})
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

func createIndividualFromIdp(email string, externalId string, organisationId string, idpId string) error {
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organisationId)

	var newIndividual individual.Individual
	newIndividual.Id = primitive.NewObjectID()
	newIndividual.Email = email
	newIndividual.ExternalId = externalId
	newIndividual.OrganisationId = organisationId
	newIndividual.IsDeleted = false
	newIndividual.IsOnboardedFromId = true
	newIndividual.IdentityProviderId = idpId

	_, err := individualRepo.Add(newIndividual)
	if err != nil {
		return err
	}
	return nil
}
