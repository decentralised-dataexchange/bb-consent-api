package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
)

// Helper function to update identity provider to iGrant.io IAM
func updateIdentityProvider(identityProviderAlias string, identityProviderRepresentation org.IdentityProviderRepresentation, adminToken string) (int, iamError, error) {
	var e iamError
	var status = http.StatusInternalServerError
	jsonReq, _ := json.Marshal(identityProviderRepresentation)
	req, err := http.NewRequest("PUT", iamConfig.URL+"/admin/realms/"+iamConfig.Realm+"/identity-provider/instances/"+identityProviderAlias, bytes.NewBuffer(jsonReq))
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
		e.ErrorType = "Identity provider update failed"
		return resp.StatusCode, e, errors.New("failed to update identity provider")
	}

	defer resp.Body.Close()

	return resp.StatusCode, e, err
}

func getClientsInRealm(clientID string, adminToken string) (string, int, iamError, error) {
	var e iamError
	var status = http.StatusInternalServerError

	req, err := http.NewRequest("GET", iamConfig.URL+"/admin/realms/"+iamConfig.Realm+"/clients?clientId="+clientID, nil)
	if err != nil {
		return "", status, e, err
	}

	req.Header.Add("Authorization", "Bearer "+adminToken)
	req.Header.Add(config.ContentTypeHeader, config.ContentTypeJSON)

	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", status, e, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		type errorMsg struct {
			ErrorMessage string `json:"errorMessage"`
		}
		var errMsg errorMsg
		json.Unmarshal(body, &errMsg)

		e.Error = errMsg.ErrorMessage
		e.ErrorType = "OpenID client secret generation failed"
		return "", resp.StatusCode, e, errors.New("failed to generate secret for OpenID client")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	var responseBody []map[string]interface{}
	json.Unmarshal(body, &responseBody)

	defer resp.Body.Close()

	return responseBody[0]["id"].(string), resp.StatusCode, e, err

}

// Helper function to update OpenID client to manage login sessions for the external identity provider
func updateOpenIDClient(clientUUID string, keycloakOpenIDClient org.KeycloakOpenIDClient, adminToken string) (int, iamError, error) {

	var e iamError
	var status = http.StatusInternalServerError
	jsonReq, _ := json.Marshal(keycloakOpenIDClient)
	req, err := http.NewRequest("PUT", iamConfig.URL+"/admin/realms/"+iamConfig.Realm+"/clients/"+clientUUID, bytes.NewBuffer(jsonReq))
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
		e.ErrorType = "OpenID client update failed"
		return resp.StatusCode, e, errors.New("failed to update OpenID client")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	var responseBody interface{}
	json.Unmarshal(body, &responseBody)

	return resp.StatusCode, e, err

}

// UpdateIdentityProvider Update external identity provider for an organisation
func UpdateIdentityProvider(w http.ResponseWriter, r *http.Request) {

	// Get the org ID and fetch the organization from the db.
	organizationID := r.Header.Get(config.OrganizationId)
	o, err := org.Get(organizationID)

	if err != nil {
		m := fmt.Sprintf("Failed to fetch org; Failed to update identity provider for %v", organizationID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	if !o.ExternalIdentityProviderAvailable {
		m := fmt.Sprintf("External IDP provider doesn't exist; Try to create instead of update; Failed to create identity provider for %v", organizationID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	// Deserializing the request payload to struct
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	var updateIdentityProviderReq IdentityProviderReq
	json.Unmarshal(b, &updateIdentityProviderReq)

	// validating request payload
	valid, err := govalidator.ValidateStruct(updateIdentityProviderReq)
	if !valid {
		m := fmt.Sprintf("Missing mandatory params for updating identity provider for org:%v\n", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	var identityProviderOpenIDConfig org.IdentityProviderOpenIDConfig

	// Update OpenID config
	identityProviderOpenIDConfig.AuthorizationURL = updateIdentityProviderReq.AuthorizationURL
	identityProviderOpenIDConfig.TokenURL = updateIdentityProviderReq.TokenURL
	identityProviderOpenIDConfig.LogoutURL = updateIdentityProviderReq.LogoutURL
	identityProviderOpenIDConfig.ClientID = updateIdentityProviderReq.ClientID
	identityProviderOpenIDConfig.ClientSecret = updateIdentityProviderReq.ClientSecret
	identityProviderOpenIDConfig.JWKSURL = updateIdentityProviderReq.JWKSURL
	identityProviderOpenIDConfig.UserInfoURL = updateIdentityProviderReq.UserInfoURL
	identityProviderOpenIDConfig.ValidateSignature = updateIdentityProviderReq.ValidateSignature
	identityProviderOpenIDConfig.DefaultScope = updateIdentityProviderReq.DefaultScope

	if len(strings.TrimSpace(updateIdentityProviderReq.LogoutURL)) > 0 {
		identityProviderOpenIDConfig.BackchannelSupported = true
	} else {
		identityProviderOpenIDConfig.BackchannelSupported = false
	}

	identityProviderOpenIDConfig.DisableUserInfo = updateIdentityProviderReq.DisableUserInfo
	identityProviderOpenIDConfig.Issuer = updateIdentityProviderReq.Issuer

	if len(strings.TrimSpace(updateIdentityProviderReq.JWKSURL)) > 0 {
		identityProviderOpenIDConfig.UseJWKSURL = true
	} else {
		identityProviderOpenIDConfig.UseJWKSURL = false
	}

	identityProviderOpenIDConfig.SyncMode = "IMPORT"
	identityProviderOpenIDConfig.ClientAuthMethod = "client_secret_post"
	identityProviderOpenIDConfig.HideOnLoginPage = true

	// Fetch admin token from keycloak
	t, status, _, err := getAdminToken()
	if err != nil {
		m := fmt.Sprintf("Failed to fetch the admin token from keycloak; Failed to update identity provider for %v", organizationID)
		common.HandleError(w, status, m, err)
		return
	}

	// Updating identity provider OpenID config
	o.IdentityProviderRepresentation.Config = identityProviderOpenIDConfig

	// Update identity provider in iGrant.io IAM
	status, _, err = updateIdentityProvider(o.IdentityProviderRepresentation.Alias, o.IdentityProviderRepresentation, t.AccessToken)
	if err != nil {
		m := fmt.Sprintf("Failed to create external identity provider in keycloak; Failed to update identity provider for %v", organizationID)
		common.HandleError(w, status, m, err)
		return
	}

	// Update the identity provider details to the database
	o, err = org.UpdateIdentityProviderByOrgID(organizationID, o.IdentityProviderRepresentation)
	if err != nil {
		m := fmt.Sprintf("Failed to update IDP config to database; Failed to update identity provider for %v", organizationID)
		common.HandleError(w, status, m, err)
		return
	}

	// FIX ME : Is this right practice to do it anonymous function executed in a separate thread ?
	go func() {
		//  Update the OpenID client
		o.KeycloakOpenIDClient.Attributes.BackchannelLogoutURL = updateIdentityProviderReq.LogoutURL

		// Fetch OpenID client UUID
		openIDClientUUID, _, _, err := getClientsInRealm(o.KeycloakOpenIDClient.ClientID, t.AccessToken)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch OpenID client UUID from keycloak; Failed to update identity provider for %v", organizationID)
			common.HandleError(w, http.StatusNotFound, m, err)
			return
		}

		// Update OpenID client to iGrant.io IAM
		status, _, err = updateOpenIDClient(openIDClientUUID, o.KeycloakOpenIDClient, t.AccessToken)
		if err != nil {
			m := fmt.Sprintf("Failed to udpate OpenID client in keycloak; Failed to update identity provider for %v", organizationID)
			log.Println(m)
			return
		}

		// Update the OpenID client details to the database
		o, err = org.UpdateOpenIDClientByOrgID(organizationID, o.KeycloakOpenIDClient)
		if err != nil {
			m := fmt.Sprintf("Failed to update OpenID client config to database; Failed to update identity provider for %v", organizationID)
			log.Println(m)
			return
		}
	}()

	response, _ := json.Marshal(o.IdentityProviderRepresentation.Config)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
