package handlerv2

import (
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	wh "github.com/bb-consent/api/src/webhooks"
	"github.com/gorilla/mux"
)

// DeleteWebhook Deletes a webhook for an organisation
func DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	// Reading URL parameters
	organizationID := r.Header.Get(config.OrganizationId)
	webhookID := mux.Vars(r)[config.WebhookId]

	// Validating the given organisation ID
	_, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	sanitizedOrgId := common.Sanitize(organizationID)
	sanitizedWebhookId := common.Sanitize(webhookID)

	// Validating the given webhook ID for an organisation
	_, err = wh.GetByOrgID(sanitizedWebhookId, sanitizedOrgId)
	if err != nil {
		m := fmt.Sprintf("Failed to get webhook:%v for organisation: %v", webhookID, organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	// Deleting webhook
	err = wh.DeleteWebhook(webhookID)
	if err != nil {
		m := fmt.Sprintf("Failed to delete webhook:%v for organisation: %v", webhookID, organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusNoContent)

}
