package middleware

import (
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/org"
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
			switch ApplicationMode {
			case config.SingleTenant:
				err := singleTenantConfig(r)
				if err != nil {
					m := "Failed to find organization"
					common.HandleError(w, http.StatusBadRequest, m, err)
					return
				}

			case config.MultiTenant:
			default:
				err := singleTenantConfig(r)
				if err != nil {
					m := "Failed to find organization"
					common.HandleError(w, http.StatusBadRequest, m, err)
					return
				}
			}

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}

// singleTenantConfig
func singleTenantConfig(r *http.Request) error {
	organization, err := org.GetFirstOrganization()
	if err != nil {
		return err
	}
	organizationId := organization.ID
	r.Header.Set(config.OrganizationId, organizationId)
	return nil
}
