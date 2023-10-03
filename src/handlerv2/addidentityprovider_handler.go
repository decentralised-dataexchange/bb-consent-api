package handlerv2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
)

var timeout time.Duration

var iamConfig config.Iam
var twilioConfig config.Twilio

// IamInit Initialize the IAM handler
func IamInit(config *config.Configuration) {
	iamConfig = config.Iam
	twilioConfig = config.Twilio
	timeout = time.Duration(time.Duration(iamConfig.Timeout) * time.Second)

	/*
		memStorage := storage.NewMemoryStorage()
		s := scheduler.New(memStorage)
		_, err := s.RunEvery(24*time.Hour, clearStaleOtps)
		if err != nil {
			log.Printf("err in scheduling clearStaleOtps: %v", err)
		}

		//TODO: Enable this later phase
		//s.Start()
	*/
}

// IdentityProviderReq Describes the request payload to create and update an external identity provider
type IdentityProviderReq struct {
	AuthorizationURL  string `json:"authorizationUrl" valid:"required"`
	TokenURL          string `json:"tokenUrl" valid:"required"`
	LogoutURL         string `json:"logoutUrl"`
	ClientID          string `json:"clientId" valid:"required"`
	ClientSecret      string `json:"clientSecret" valid:"required"`
	JWKSURL           string `json:"jwksUrl"`
	UserInfoURL       string `json:"userInfoUrl"`
	ValidateSignature bool   `json:"validateSignature"`
	DisableUserInfo   bool   `json:"disableUserInfo"`
	Issuer            string `json:"issuer"`
	DefaultScope      string `json:"defaultScope"`
}

type iamToken struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
}

type iamError struct {
	ErrorType string `json:"error"`
	Error     string `json:"error_description"`
}

func getAdminToken() (iamToken, int, iamError, error) {
	t, status, iamErr, err := getToken(iamConfig.AdminUser, iamConfig.AdminPassword, "admin-cli", "master")
	return t, status, iamErr, err
}

func getToken(username string, password string, clientID string, realm string) (iamToken, int, iamError, error) {
	var tok iamToken
	var e iamError
	var status = http.StatusInternalServerError

	data := url.Values{}
	data.Set("username", username)
	data.Add("password", password)
	data.Add("client_id", clientID)
	data.Add("grant_type", "password")

	resp, err := http.PostForm(iamConfig.URL+"/realms/"+realm+"/protocol/openid-connect/token", data)
	if err != nil {
		return tok, status, e, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return tok, status, e, err
	}
	if resp.StatusCode != http.StatusOK {
		var e iamError
		json.Unmarshal(body, &e)
		return tok, resp.StatusCode, e, errors.New("failed to get token")
	}
	json.Unmarshal(body, &tok)

	return tok, resp.StatusCode, e, err
}

// Helper function to add identity provider to iGrant.io IAM
func addIdentityProvider(identityProviderRepresentation org.IdentityProviderRepresentation, adminToken string) (int, iamError, error) {
	var e iamError
	var status = http.StatusInternalServerError
	jsonReq, _ := json.Marshal(identityProviderRepresentation)
	req, err := http.NewRequest("POST", iamConfig.URL+"/admin/realms/"+iamConfig.Realm+"/identity-provider/instances", bytes.NewBuffer(jsonReq))
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

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		type errorMsg struct {
			ErrorMessage string `json:"errorMessage"`
		}
		var errMsg errorMsg
		json.Unmarshal(body, &errMsg)
		e.Error = errMsg.ErrorMessage
		e.ErrorType = "Identity provider creation failed"
		return resp.StatusCode, e, errors.New("Failed to create identity provider")
	}

	defer resp.Body.Close()

	return resp.StatusCode, e, err
}

// Helper function to add OpenID client to manage login sessions for the external identity provider
func addOpenIDClient(keycloakOpenIDClient org.KeycloakOpenIDClient, adminToken string) (int, iamError, error) {

	var e iamError
	var status = http.StatusInternalServerError
	jsonReq, _ := json.Marshal(keycloakOpenIDClient)
	req, err := http.NewRequest("POST", iamConfig.URL+"/admin/realms/"+iamConfig.Realm+"/clients", bytes.NewBuffer(jsonReq))
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

	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		type errorMsg struct {
			ErrorMessage string `json:"errorMessage"`
		}
		var errMsg errorMsg
		json.Unmarshal(body, &errMsg)
		e.Error = errMsg.ErrorMessage
		e.ErrorType = "OpenID client creation failed"
		return resp.StatusCode, e, errors.New("Failed to create OpenID client")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	var responseBody interface{}
	json.Unmarshal(body, &responseBody)

	return resp.StatusCode, e, err

}

// AddIdentityProvider Add an external identity provider for an organization
func AddIdentityProvider(w http.ResponseWriter, r *http.Request) {

	// Note: Set OpenID-Connect as subscription method for the organization
	//       Execute set subscription method API to do the same.

	// Get the org ID and fetch the organization from the db.
	organizationID := r.Header.Get(config.OrganizationId)
	o, err := org.Get(organizationID)

	if err != nil {
		m := fmt.Sprintf("Failed to fetch org; Failed to create identity provider for %v", organizationID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	if o.ExternalIdentityProviderAvailable {
		m := fmt.Sprintf("External IDP exists; Try to update instead of create; Failed to create identity provider for %v", organizationID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	// Deserializing the request payload to struct
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	var addIdentityProviderReq IdentityProviderReq
	json.Unmarshal(b, &addIdentityProviderReq)

	// validating request payload
	valid, err := govalidator.ValidateStruct(addIdentityProviderReq)
	if valid != true {
		m := fmt.Sprintf("Missing mandatory params for creating identity provider for org:%v\n", organizationID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	var identityProviderOpenIDConfig org.IdentityProviderOpenIDConfig

	// OpenID config
	identityProviderOpenIDConfig.AuthorizationURL = addIdentityProviderReq.AuthorizationURL
	identityProviderOpenIDConfig.TokenURL = addIdentityProviderReq.TokenURL
	identityProviderOpenIDConfig.LogoutURL = addIdentityProviderReq.LogoutURL
	identityProviderOpenIDConfig.ClientID = addIdentityProviderReq.ClientID
	identityProviderOpenIDConfig.ClientSecret = addIdentityProviderReq.ClientSecret
	identityProviderOpenIDConfig.JWKSURL = addIdentityProviderReq.JWKSURL
	identityProviderOpenIDConfig.UserInfoURL = addIdentityProviderReq.UserInfoURL
	identityProviderOpenIDConfig.ValidateSignature = addIdentityProviderReq.ValidateSignature
	identityProviderOpenIDConfig.DefaultScope = addIdentityProviderReq.DefaultScope

	if len(strings.TrimSpace(addIdentityProviderReq.LogoutURL)) > 0 {
		identityProviderOpenIDConfig.BackchannelSupported = true
	} else {
		identityProviderOpenIDConfig.BackchannelSupported = false
	}

	identityProviderOpenIDConfig.DisableUserInfo = addIdentityProviderReq.DisableUserInfo
	identityProviderOpenIDConfig.Issuer = addIdentityProviderReq.Issuer

	if len(strings.TrimSpace(addIdentityProviderReq.JWKSURL)) > 0 {
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
		m := fmt.Sprintf("Failed to fetch the admin token from keycloak; Failed to create identity provider for %v", organizationID)
		common.HandleError(w, status, m, err)
		return
	}

	// Constructing the request payload for creating identity provider
	var identityProviderRepresentation = org.IdentityProviderRepresentation{
		Config:                    identityProviderOpenIDConfig,
		Alias:                     o.ID.Hex(),
		ProviderID:                "oidc",
		AuthenticateByDefault:     false,
		Enabled:                   true,
		FirstBrokerLoginFlowAlias: iamConfig.ExternalIdentityProvidersConfiguration.IdentityProviderCustomerAutoLinkFlowName,
		LinkOnly:                  false,
		AddReadTokenRoleOnCreate:  false,
		PostBrokerLoginFlowAlias:  "",
		StoreToken:                false,
		TrustEmail:                false,
	}

	// Add identity provider to iGrant.io IAM
	status, _, err = addIdentityProvider(identityProviderRepresentation, t.AccessToken)
	if err != nil {
		m := fmt.Sprintf("Failed to create external identity provider in keycloak; Failed to create identity provider for %v", organizationID)
		common.HandleError(w, status, m, err)
		return
	}

	// Save the identity provider details to the database
	o, err = org.UpdateIdentityProviderByOrgID(organizationID, identityProviderRepresentation)
	if err != nil {
		m := fmt.Sprintf("Failed to update IDP config to database; Failed to create identity provider for %v", organizationID)
		common.HandleError(w, status, m, err)
		return
	}

	// Update external identity provider available status
	o, err = org.UpdateExternalIdentityProviderAvailableStatus(organizationID, true)
	if err != nil {
		m := fmt.Sprintf("Failed to update external identity provider available status; Failed to create identity provider for %v", organizationID)
		common.HandleError(w, status, m, err)
		return
	}

	// FIX ME : Is this right practice to do it anonymous function executed in a separate thread ?
	go func() {
		// Create a client to manage the user session from external identity provider

		// ID for a custom authentication flow created in the keycloak to manage the conversion of external access token to internal (iGrant.io) authorization code
		var IDPCustomKeycloakAuthenticationFlow = iamConfig.ExternalIdentityProvidersConfiguration.IdentityProviderCustomerAuthenticationFlowID

		// Construct the request payload to create OpenID client
		var keycloakOpenIDClient = org.KeycloakOpenIDClient{
			Access: org.KeycloakOpenIDClientAccess{
				Configure: true,
				View:      true,
				Manage:    true,
			},
			AlwaysDisplayInConsole: false,
			Attributes: org.KeycloakOpenIDClientAttributes{
				BackchannelLogoutRevokeOfflineTokens:  "true",
				BackchannelLogoutSessionRequired:      "true",
				BackchannelLogoutURL:                  identityProviderOpenIDConfig.LogoutURL,
				ClientCredentialsUseRefreshToken:      "false",
				DisplayOnConsentScreen:                "false",
				ExcludeSessionStateFromAuthResponse:   "false",
				SamlAssertionSignature:                "false",
				SamlAuthnstatement:                    "false",
				SamlClientSignature:                   "false",
				SamlEncrypt:                           "false",
				SamlForcePostBinding:                  "false",
				SamlMultivaluedRoles:                  "false",
				SamlOnetimeuseCondition:               "false",
				SamlServerSignature:                   "false",
				SamlServerSignatureKeyinfoExt:         "false",
				SamlForceNameIDFormat:                 "false",
				TLSClientCertificateBoundAccessTokens: "false",
			},
			AuthenticationFlowBindingOverrides: org.KeycloakOpenIDClientAuthenticationFlowBindingOverrides{
				Browser:     IDPCustomKeycloakAuthenticationFlow,
				DirectGrant: IDPCustomKeycloakAuthenticationFlow,
			},
			BearerOnly:              false,
			ClientAuthenticatorType: "client-secret",
			ClientID:                o.ID.Hex(),
			ConsentRequired:         false,
			DefaultClientScopes: []string{
				"web-origins",
				"role_list",
				"profile",
				"roles",
				"email",
			},
			DirectAccessGrantsEnabled: true,
			Enabled:                   true,
			FrontchannelLogout:        false,
			FullScopeAllowed:          true,
			ImplicitFlowEnabled:       false,
			NodeReRegistrationTimeout: -1,
			NotBefore:                 0,
			OptionalClientScopes: []string{
				"address",
				"phone",
				"offline_access",
				"microprofile-jwt",
			},
			Protocol:               "openid-connect",
			PublicClient:           false,
			RedirectUris:           []string{},
			ServiceAccountsEnabled: false,
			StandardFlowEnabled:    true,
			SurrogateAuthRequired:  false,
			WebOrigins:             []string{},
		}

		// Add OpenID client to iGrant.io IAM
		status, _, err = addOpenIDClient(keycloakOpenIDClient, t.AccessToken)
		if err != nil {
			m := fmt.Sprintf("Failed to create OpenID client in keycloak; Failed to create identity provider for %v", organizationID)
			log.Println(m)
			return
		}

		// Save the OpenID client details to the database
		o, err = org.UpdateOpenIDClientByOrgID(organizationID, keycloakOpenIDClient)
		if err != nil {
			m := fmt.Sprintf("Failed to update OpenID client config to database; Failed to create identity provider for %v", organizationID)
			log.Println(m)
			return
		}

	}()

	response, _ := json.Marshal(o.IdentityProviderRepresentation.Config)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(response)

}
