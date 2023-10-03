package handlerv2

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
)

// Helper function to delete identity provider to iGrant.io IAM
func deleteIdentityProvider(identityProviderAlias string, adminToken string) (int, iamError, error) {
	var e iamError
	var status = http.StatusInternalServerError
	req, err := http.NewRequest("DELETE", iamConfig.URL+"/admin/realms/"+iamConfig.Realm+"/identity-provider/instances/"+identityProviderAlias, nil)
	if err != nil {
		return status, e, err
	}

	req.Header.Add("Authorization", "Bearer "+adminToken)
	req.Header.Add(config.ContentTypeHeader, config.ContentTypeJSON)

	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return status, e, err
	}

	if resp.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		type errorMsg struct {
			ErrorMessage string `json:"errorMessage"`
		}
		var errMsg errorMsg
		json.Unmarshal(body, &errMsg)
		e.Error = errMsg.ErrorMessage
		e.ErrorType = "Identity provider delete failed"
		return resp.StatusCode, e, errors.New("failed to delete identity provider")
	}

	defer resp.Body.Close()

	return resp.StatusCode, e, err
}

// Helper function to delete OpenID client to manage login sessions for the external identity provider
func deleteOpenIDClient(clientUUID string, adminToken string) (int, iamError, error) {

	var e iamError
	var status = http.StatusInternalServerError
	req, err := http.NewRequest("DELETE", iamConfig.URL+"/admin/realms/"+iamConfig.Realm+"/clients/"+clientUUID, nil)
	if err != nil {
		return status, e, err
	}

	req.Header.Add("Authorization", "Bearer "+adminToken)
	req.Header.Add(config.ContentTypeHeader, config.ContentTypeJSON)

	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return status, e, err
	}

	if resp.StatusCode != http.StatusNoContent {
		body, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		type errorMsg struct {
			ErrorMessage string `json:"errorMessage"`
		}
		var errMsg errorMsg
		json.Unmarshal(body, &errMsg)
		e.Error = errMsg.ErrorMessage
		e.ErrorType = "OpenID client delete failed"
		return resp.StatusCode, e, errors.New("failed to delete OpenID client")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	var responseBody interface{}
	json.Unmarshal(body, &responseBody)

	return resp.StatusCode, e, err

}

// DeleteIdentityProvider Delete external identity provider for an organisation
func DeleteIdentityProvider(w http.ResponseWriter, r *http.Request) {

	// Get the org ID and fetch the organization from the db.
	organizationID := r.Header.Get(config.OrganizationId)
	o, err := org.Get(organizationID)

	if err != nil {
		m := fmt.Sprintf("Failed to fetch org; Failed to delete identity provider for %v", organizationID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	if !o.ExternalIdentityProviderAvailable {
		m := fmt.Sprintf("External IDP provider doesn't exist; Try to create instead of delete; Failed to delete identity provider for %v", organizationID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	// Fetch admin token from keycloak
	t, status, _, err := getAdminToken()
	if err != nil {
		m := fmt.Sprintf("Failed to fetch the admin token from keycloak; Failed to delete identity provider for %v", organizationID)
		common.HandleError(w, status, m, err)
		return
	}

	// Delete identity provider in IAM
	status, _, err = deleteIdentityProvider(o.IdentityProviderRepresentation.Alias, t.AccessToken)
	if err != nil {
		m := fmt.Sprintf("Failed to delete external identity provider in keycloak; Failed to delete identity provider for %v", organizationID)
		common.HandleError(w, status, m, err)
		return
	}

	// Update external identity provider available status
	o, err = org.UpdateExternalIdentityProviderAvailableStatus(organizationID, false)
	if err != nil {
		m := fmt.Sprintf("Failed to update external identity provider available status; Failed to delete identity provider for %v", organizationID)
		common.HandleError(w, status, m, err)
		return
	}

	// Delete the identity provider details from the database
	o, err = org.DeleteIdentityProviderByOrgID(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to delete IDP config to database; Failed to delete identity provider for %v", organizationID)
		common.HandleError(w, status, m, err)
		return
	}

	// FIX ME : Is this right practice to do it anonymous function executed in a separate thread ?
	go func() {

		// Fetch OpenID client UUID from IAM
		openIDClientUUID, _, _, err := getClientsInRealm(o.KeycloakOpenIDClient.ClientID, t.AccessToken)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch OpenID client UUID from keycloak; Failed to delete identity provider for %v", organizationID)
			log.Println(m)
			return
		}

		// Delete OpenID client in iGrant.io IAM
		_, _, err = deleteOpenIDClient(openIDClientUUID, t.AccessToken)
		if err != nil {
			m := fmt.Sprintf("Failed to delete external identity provider in keycloak; Failed to delete identity provider for %v", organizationID)
			log.Println(m)
			return
		}

		// Delete the OpenID client details from the database
		_, err = org.DeleteOpenIDClientByOrgID(organizationID)
		if err != nil {
			m := fmt.Sprintf("Failed to delete OpenID config to database; Failed to delete identity provider for %v", organizationID)
			log.Println(m)
			return
		}

	}()

	w.WriteHeader(http.StatusNoContent)
	w.Write(nil)
}
