package middleware

import (
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/actionlog"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/token"
)

// LogApiCalls Logs API(s).
func LogApiCalls() Middleware {
	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {

			organisationId := r.Header.Get(config.OrganizationId)
			aLog := fmt.Sprintf("%v: %v called by user: %v", r.Method, r.URL.Path, token.GetUserID(r))
			actionlog.LogOrgAPICalls(token.GetUserID(r), token.GetUserName(r), organisationId, aLog)

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}
