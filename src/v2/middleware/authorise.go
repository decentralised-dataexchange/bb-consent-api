package middleware

import (
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/v2/token"
	"github.com/casbin/casbin/v2"
)

func Authorize(e *casbin.Enforcer) Middleware {

	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {

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

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}
