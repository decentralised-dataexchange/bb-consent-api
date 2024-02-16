package middleware

import (
	"net/http"
	"strings"

	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/error_handler"
	"github.com/bb-consent/api/internal/token"
	"github.com/gorilla/mux"
)

// ValidateIndividualId Validates the individual id in path variable.
func ValidateIndividualId() Middleware {

	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {
			// To catch panic and recover the error
			// Once the error is recovered respond by
			// writing the error to HTTP response
			defer error_handler.HandleExit(w)
			headerType, _ := getAccessTokenFromHeader(w, r)

			if headerType == token.AuthorizationToken {
				// Get the path parameters from the request
				vars := mux.Vars(r)

				// Check if "individualId" is present in the path parameters
				if individualId, ok := vars[config.IndividualId]; ok {
					// Process the request with the individualId path parameter
					requestedIndividualId := token.GetUserID(r)

					if strings.TrimSpace(individualId) != strings.TrimSpace(requestedIndividualId) {
						m := "Unauthorized access;User doesn't have enough permissions;"
						error_handler.Exit(http.StatusBadRequest, m)
					}
				}
			}

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}
