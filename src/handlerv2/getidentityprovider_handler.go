package handlerv2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
)

// GetIdentityProvider Get external identity provider for an organisation
func GetIdentityProvider(w http.ResponseWriter, r *http.Request) {

	// Get the org ID and fetch the organization from the db.
	organizationID := r.Header.Get(config.OrganizationId)
	o, err := org.Get(organizationID)

	if err != nil {
		m := fmt.Sprintf("Failed to fetch org; Failed to get identity provider for %v", organizationID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	if !o.ExternalIdentityProviderAvailable {
		m := fmt.Sprintf("External IDP provider doesn't exist; Try to create instead of get; Failed to get identity provider for %v", organizationID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	response, _ := json.Marshal(o.IdentityProviderRepresentation.Config)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
