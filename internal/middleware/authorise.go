package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/token"
	"github.com/casbin/casbin/v2"
)

func Authorize(e *casbin.Enforcer) Middleware {

	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {

			headerType, headerValue := getAccessTokenFromHeader(w, r)

			// verify rbac for token based access
			if headerType == token.AuthorizationToken {
				userRole := token.GetUserRole(r)

				// casbin enforce
				res, err := e.Enforce(userRole, r.URL.Path, r.Method)
				if err != nil {
					m := "Failed to enforce casbin authentication;"
					common.HandleError(w, http.StatusInternalServerError, m, err)
					return
				}

				if !res {
					log.Printf("User does not have enough permissions")
					m := "Unauthorized access;User doesn't have enough permissions;"
					common.HandleError(w, http.StatusForbidden, m, nil)
					return
				}
			}
			// verify rbac for apikey based access
			if headerType == token.AuthorizationAPIKey {
				// decode claims
				claims := decodeApiKey(headerValue, w)
				res, err := verifyApiKeyScope(claims.Scopes, e, r)
				if err != nil {
					m := "Failed to enforce casbin authentication;"
					common.HandleError(w, http.StatusInternalServerError, m, err)
					return
				}
				if !res {
					log.Printf("User does not have enough permissions")
					m := "Unauthorized access;User doesn't have enough permissions;"
					common.HandleError(w, http.StatusForbidden, m, nil)
					return
				}
			}

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}

// verifyApiKeyScope verify apikey scope
func verifyApiKeyScope(Scopes []string, e *casbin.Enforcer, r *http.Request) (bool, error) {
	var res bool
	var err error
	for _, scope := range Scopes {
		res, err = e.Enforce(scope, r.URL.Path, r.Method)
		if err != nil {
			m := "failed to enforce casbin authentication;"
			return false, errors.New(m)
		}
		if res {
			return true, nil
		}
	}
	return false, nil
}
