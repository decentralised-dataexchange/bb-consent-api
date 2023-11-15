package middleware

import (
	"net/http"
	"strings"

	"github.com/bb-consent/api/internal/apikey"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/error_handler"
	"github.com/bb-consent/api/internal/individual"
	"github.com/bb-consent/api/internal/rbac"
	"github.com/bb-consent/api/internal/token"
	"github.com/bb-consent/api/internal/user"
)

func decodeApiKey(headerValue string, w http.ResponseWriter) apikey.Claims {
	claims, err := apikey.Decode(headerValue)

	if err != nil {
		m := "Invalid token, Authorization failed"
		error_handler.Exit(http.StatusUnauthorized, m)
	}

	return claims
}

func performAPIKeyAuthentication(claims apikey.Claims, tag string, w http.ResponseWriter, r *http.Request) {

	t := token.AccessToken{}

	// Check if individualId is present in request header for service tag
	// If present validate user
	// If not present, return error
	if tag == config.Service {
		individualId := r.Header.Get(config.IndividualHeaderKey)

		// Repository
		individualRepo := individual.IndividualRepository{}
		individualRepo.Init(claims.OrganisationId)

		if len(strings.TrimSpace(individualId)) != 0 {
			// fetch the individual
			individual, err := individualRepo.Get(individualId)
			if err != nil {
				m := "User does not exist, Authorization failed"
				error_handler.Exit(http.StatusBadRequest, m)
			}
			t.Email = individual.Email
			t.IamID = individual.IamId
			token.Set(r, t)
			token.SetUserToRequestContext(r, individualId, rbac.ROLE_USER)
		} else {
			m := "IndividualId is not present in request header"
			error_handler.Exit(http.StatusBadRequest, m)
		}

	} else {
		// fetch organisation admin and set to context if api tag is other than service
		orgAdmin, err := user.Get(claims.OrganisationAdminId)
		if err != nil {
			m := "User does not exist, Authorization failed"
			error_handler.Exit(http.StatusBadRequest, m)
		}
		t.Email = orgAdmin.Email
		t.IamID = orgAdmin.IamID
		token.Set(r, t)
		token.SetUserToRequestContext(r, claims.OrganisationAdminId, rbac.ROLE_ADMIN)
	}

}

// getApiTag get api tag from route
func getApiTag(route string) string {

	if strings.HasPrefix(route, "/v2/service") {
		return "service"
	} else if strings.HasPrefix(route, "/v2/config") {
		return "config"
	} else if strings.HasPrefix(route, "/v2/audit") {
		return "audit"
	} else if strings.HasPrefix(route, "/v2/onboard") || strings.HasPrefix(route, "/onboard") {
		return "onboard"
	} else {
		return "unknown"
	}
}

// ValidateAPIKeyAndIndividualId Validates the apikey.
func ValidateAPIKeyAndIndividualId() Middleware {

	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {
			// To catch panic and recover the error
			// Once the error is recovered respond by
			// writing the error to HTTP response
			defer error_handler.HandleExit(w)
			headerType, headerValue := getAccessTokenFromHeader(w, r)

			if headerType == token.AuthorizationAPIKey {
				claims := decodeApiKey(headerValue, w)
				tag := getApiTag(r.URL.Path)
				if tag == "unknown" {
					m := "Unknown tag in request path"
					error_handler.Exit(http.StatusBadRequest, m)
				}
				performAPIKeyAuthentication(claims, tag, w, r)
			}

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}
