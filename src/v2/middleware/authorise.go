package middleware

import (
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/rbac"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/user"
	"github.com/casbin/casbin/v2"
)

func Authorize(e *casbin.Enforcer) Middleware {

	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {

			userID := token.GetUserID(r)

			user, err := user.Get(userID)
			if err != nil {
				m := fmt.Sprintf("Failed to locate user with ID: %v", userID)
				common.HandleError(w, http.StatusBadRequest, m, err)
				return
			}
			roles := user.Roles

			var role string

			if len(roles) > 0 {
				role = rbac.ROLE_ADMIN
			} else {
				role = rbac.ROLE_USER
			}

			// casbin enforce
			res, err := e.Enforce(role, r.URL.Path, r.Method)
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
