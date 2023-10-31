package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/bb-consent/api/internal/apikey"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/error_handler"
	"github.com/bb-consent/api/internal/iam"
	"github.com/bb-consent/api/internal/idp"
	"github.com/bb-consent/api/internal/individual"
	"github.com/bb-consent/api/internal/org"
	"github.com/bb-consent/api/internal/rbac"
	"github.com/bb-consent/api/internal/token"
	"github.com/bb-consent/api/internal/user"
	"github.com/coreos/go-oidc/v3/oidc"
)

func getAccessTokenFromHeader(w http.ResponseWriter, r *http.Request) (headerType int, headerValue string) {
	headerType, headerValue, err := token.DecodeAuthHeader(r)
	if err != nil {
		m := "Invalid authorization header, Authorization failed"
		error_handler.Exit(http.StatusUnauthorized, m)
	}

	return headerType, headerValue
}

func storeAccessTokenInRequestContext(headerValue string, w http.ResponseWriter, r *http.Request) {

	t, err := token.ParseToken(headerValue)
	if err != nil {
		m := "Invalid token, Authorization failed"
		error_handler.Exit(http.StatusUnauthorized, m)
	}
	token.Set(r, t)
}

func verifyTokenAndIdentifyRole(accessToken string, r *http.Request) error {
	// Verify token against Consent BB IDP
	consentBBIssuerUrl := iam.IamConfig.URL + "/realms/" + iam.IamConfig.Realm
	consentBBJwksUrl := iam.IamConfig.URL + "/realms/" + iam.IamConfig.Realm + "/protocol/openid-connect/certs"
	jwks := oidc.NewRemoteKeySet(context.Background(), consentBBJwksUrl)
	c := oidc.NewVerifier(consentBBIssuerUrl, jwks, &oidc.Config{SkipClientIDCheck: true})
	tokenPayload, err := c.Verify(context.Background(), accessToken)

	// Get organisation
	organization, err := org.GetFirstOrganization()
	if err != nil {
		m := "Failed to fetch organisation"
		error_handler.Exit(http.StatusInternalServerError, m)
	}

	// Repository
	individualRepo := individual.IndividualRepository{}
	individualRepo.Init(organization.ID.Hex())

	if err != nil {
		// Individual doesn't belong to Consent BB IDP
		// Therefore needs to verify whether the user
		// belongs to organisation's configured IDP.

		// Repository
		idpRepo := idp.IdentityProviderRepository{}
		idpRepo.Init(organization.ID.Hex())

		// Fetch IDP for the org
		idp, err := idpRepo.GetByOrgId()
		if err != nil {
			m := "Failed to fetch idp by organisation id"
			error_handler.Exit(http.StatusInternalServerError, m)
		}

		// Verify token against organisation's IDP
		jwks := oidc.NewRemoteKeySet(context.Background(), idp.JWKSURL)
		c := oidc.NewVerifier(idp.IssuerUrl, jwks, &oidc.Config{SkipClientIDCheck: true})
		tokenPayload, err := c.Verify(context.Background(), accessToken)
		if err != nil {
			// If token verification failed, then the user doesn't belong to
			// Consent BB IDP or organisation's IDP
			m := "Failed to verify token"
			error_handler.Exit(http.StatusUnauthorized, m)
		}

		// Query individual by `externalId` to
		// check if an existing individuals is present.
		externalId := tokenPayload.Subject
		individual, err := individualRepo.GetByExternalId(externalId)
		if err != nil {
			m := "User does not exist, Authorization failed"
			error_handler.Exit(http.StatusBadRequest, m)
		}

		// Set user Id and user roles to request context
		token.SetUserToRequestContext(r, individual.Id.Hex(), rbac.ROLE_USER)

		return nil

	}

	// If individual is present in Consent BB IDP
	// Query by `iamId` and fetch individual
	iamId := tokenPayload.Subject
	user, err := user.GetByIamID(iamId)
	if err != nil {

		// Get individual
		individual, err := individualRepo.Get(iamId)
		if err != nil {
			m := "User does not exist, Authorization failed"
			error_handler.Exit(http.StatusBadRequest, m)
		}

		// Set user Id and user roles to request context
		token.SetUserToRequestContext(r, individual.Id.Hex(), rbac.ROLE_USER)
	}

	// Set user Id and user roles to request context
	if len(user.Roles) > 0 {
		token.SetUserToRequestContext(r, user.ID.Hex(), rbac.ROLE_ADMIN)
	} else {
		token.SetUserToRequestContext(r, user.ID.Hex(), rbac.ROLE_USER)
	}

	return nil
}

func decodeApiKey(headerValue string, w http.ResponseWriter) apikey.Claims {
	claims, err := apikey.Decode(headerValue)

	if err != nil {
		m := "Invalid token, Authorization failed"
		error_handler.Exit(http.StatusUnauthorized, m)
	}

	return claims
}

func performAPIKeyAuthentication(claims apikey.Claims, w http.ResponseWriter, r *http.Request) {
	individualId := r.Header.Get(config.IndividualHeaderKey)

	t := token.AccessToken{}
	token.Set(r, t)
	if len(strings.TrimSpace(individualId)) != 0 {
		token.SetUserToRequestContext(r, individualId, rbac.ROLE_USER)
	} else {
		token.SetUserToRequestContext(r, claims.OrganisationAdminId, rbac.ROLE_ADMIN)
	}
}

// Authenticate Validates the token and sets the token to the context.
func Authenticate() Middleware {

	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {
			// To catch panic and recover the error
			// Once the error is recovered respond by
			// writing the error to HTTP response
			defer error_handler.HandleExit(w)
			headerType, headerValue := getAccessTokenFromHeader(w, r)

			if headerType == token.AuthorizationToken {
				storeAccessTokenInRequestContext(headerValue, w, r)
				verifyTokenAndIdentifyRole(headerValue, r)
			}
			if headerType == token.AuthorizationAPIKey {
				claims := decodeApiKey(headerValue, w)
				performAPIKeyAuthentication(claims, w, r)
			}
			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}
