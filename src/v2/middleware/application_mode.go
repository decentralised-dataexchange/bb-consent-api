package middleware

import (
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
)

var ApplicationMode string
var Organization config.Organization

func ApplicationModeInit(config *config.Configuration) {
	ApplicationMode = config.ApplicationMode
	Organization = config.Organization
}

// SetApplicationMode sets application modes for routes to either single tenant or multi tenant
func SetApplicationMode() Middleware {
	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {

			if ApplicationMode == config.SingleTenant {

				organization, err := org.GetFirstOrganization()
				if err != nil {
					m := "failed to find organization"
					common.HandleError(w, http.StatusBadRequest, m, err)
					return
				}
				organizationId := organization.ID.Hex()
				r.Header.Set(config.OrganizationId, organizationId)
			}

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}
