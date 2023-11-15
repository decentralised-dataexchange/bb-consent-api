package middleware

import (
	"net/http"

	"github.com/bb-consent/api/internal/apikey"
	"github.com/bb-consent/api/internal/error_handler"
	"github.com/bb-consent/api/internal/rbac"
	"github.com/bb-consent/api/internal/token"
	"github.com/bb-consent/api/internal/user"
)

func apiKeyAuthentication(claims apikey.Claims, w http.ResponseWriter, r *http.Request) {

	t := token.AccessToken{}

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

// ValidateAPIKey Validates the apikey.
func ValidateAPIKey() Middleware {

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
				apiKeyAuthentication(claims, w, r)
			}

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}
