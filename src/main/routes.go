package main

import (
	"net/http"

	"github.com/bb-consent/api/src/handler"
	m "github.com/bb-consent/api/src/middleware"
	"github.com/gorilla/mux"
)

// Root access return 200 OK for health check when the api
// is deployed in K8s with ingress controller.
func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

// SetRoutes sets the routes that the back end server serves
func SetRoutes(r *mux.Router) {
	r.HandleFunc("/", healthz).Methods("GET")

	r.Handle("/v1/organizations", m.Chain(handler.AddOrganization, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/roles", m.Chain(handler.GetOrganizationRoles, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/subscribe-methods", m.Chain(handler.GetSubscribeMethods, m.Logger(), m.Authenticate())).Methods("GET")

	r.Handle("/v1/organizations/data-requests", m.Chain(handler.GetDataRequestStatus, m.Logger(), m.Authenticate())).Methods("GET")

	r.Handle("/v1/organizations/types", m.Chain(handler.GetOrganizationTypes, m.LoggerNoAuth())).Methods("GET")
	r.Handle("/v1/organizations/types", m.Chain(handler.AddOrganizationType, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/types/{typeID}", m.Chain(handler.UpdateOrganizationType, m.Logger(), m.Authenticate())).Methods("PATCH")
	r.Handle("/v1/organizations/types/{typeID}", m.Chain(handler.DeleteOrganizationType, m.Logger(), m.Authenticate())).Methods("DELETE")
	r.Handle("/v1/organizations/types/{typeID}", m.Chain(handler.GetOrganizationTypeByID, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/types/{typeID}/image", m.Chain(handler.UpdateOrganizationTypeImage, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/types/{typeID}/image", m.Chain(handler.GetOrganizationTypeImage, m.Logger(), m.Authenticate())).Methods("GET")

	// Organization webhook event types
	r.Handle("/v1/organizations/webhooks/event-types", m.Chain(handler.GetWebhookEventTypes, m.Logger(), m.Authenticate())).Methods("GET")

	r.Handle("/v1/organizations/{organizationID}", m.Chain(handler.GetOrganizationByID, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{organizationID}", m.Chain(handler.UpdateOrganization, m.Logger(), m.Authenticate())).Methods("PATCH")
	r.Handle("/v1/organizations/{organizationID}/coverimage", m.Chain(handler.UpdateOrganizationCoverImage, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/logoimage", m.Chain(handler.UpdateOrganizationLogoImage, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/image/{imageID}", m.Chain(handler.GetOrganizationImage, m.Logger(), m.Authenticate())).Methods("GET")

	r.Handle("/v1/organizations/{organizationID}/eulaURL", m.Chain(handler.UpdateOrgEula, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/eulaURL", m.Chain(handler.DeleteOrgEula, m.Logger(), m.Authenticate())).Methods("DELETE")

	r.Handle("/v1/organizations/{organizationID}/admins", m.Chain(handler.AddOrgAdmin, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/admins", m.Chain(handler.GetOrgAdmins, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{organizationID}/admins", m.Chain(handler.DeleteOrgAdmin, m.Logger(), m.Authenticate())).Methods("DELETE")

	r.Handle("/v1/organizations/{organizationID}/purposes", m.Chain(handler.AddConsentPurposes, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/purposes", m.Chain(handler.GetPurposes, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{organizationID}/purposes/{purposeID}", m.Chain(handler.DeleteConsentPurposeByID, m.Logger(), m.Authenticate())).Methods("DELETE")
	r.Handle("/v1/organizations/{organizationID}/purposes/{purposeID}", m.Chain(handler.UpdatePurposeByID, m.Logger(), m.Authenticate())).Methods("PUT")
	r.Handle("/v1/organizations/{organizationID}/purposes/{purposeID}", m.Chain(handler.GetPurposeByID, m.Logger(), m.Authenticate())).Methods("GET")

	r.Handle("/v1/organizations/{organizationID}/templates", m.Chain(handler.AddConsentTemplates, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/templates", m.Chain(handler.GetTemplates, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{organizationID}/templates/{templateID}", m.Chain(handler.DeleteConsentTemplateByID, m.Logger(), m.Authenticate())).Methods("DELETE")
	r.Handle("/v1/organizations/{organizationID}/purposes/{purposeID}/templates", m.Chain(handler.DeleteConsentTemplatesByID, m.Logger(), m.Authenticate())).Methods("DELETE")
	r.Handle("/v1/organizations/{organizationID}/templates/{templateID}", m.Chain(handler.GetTemplateByID, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{organizationID}/templates/{templateID}", m.Chain(handler.UpdateTemplateByID, m.Logger(), m.Authenticate())).Methods("PUT")

	r.Handle("/v1/organizations/{organizationID}/users", m.Chain(handler.AddUserToOrganization, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/users/{userID}", m.Chain(handler.DeleteUserFromOrganization, m.Logger(), m.Authenticate())).Methods("DELETE")
	r.Handle("/v1/organizations/{organizationID}/users", m.Chain(handler.GetOrganizationUsers, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{organizationID}/users/count", m.Chain(handler.GetOrganizationUsersCount, m.Logger(), m.Authenticate())).Methods("GET")

	// Organization global policy configuration
	r.Handle("/v1/organizations/{organizationID}/global-policy-configuration", m.Chain(handler.GetGlobalPolicyConfiguration, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{organizationID}/global-policy-configuration", m.Chain(handler.UpdateGlobalPolicyConfiguration, m.Logger(), m.Authenticate())).Methods("POST")

	r.Handle("/v1/organizations/{organizationID}/subscription/enable", m.Chain(handler.EnableOrganizationSubscription, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/subscription/disable", m.Chain(handler.DisableOrganizationSubscription, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/subscribe-method", m.Chain(handler.GetSubscribeMethod, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{organizationID}/subscribe-method", m.Chain(handler.SetSubscribeMethod, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/subscribe-key", m.Chain(handler.GetSubscribeKey, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{organizationID}/subscribe-key/renew", m.Chain(handler.RenewSubscribeKey, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/subscription", m.Chain(handler.GetOrganizationSubscriptionStatus, m.Logger(), m.Authenticate())).Methods("GET")

	// Organisation identity provider related API(s)
	r.Handle("/v1/organizations/{organizationID}/idp/open-id", m.Chain(handler.AddIdentityProvider, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/idp/open-id", m.Chain(handler.UpdateIdentityProvider, m.Logger(), m.Authenticate())).Methods("PUT")
	r.Handle("/v1/organizations/{organizationID}/idp/open-id", m.Chain(handler.DeleteIdentityProvider, m.Logger(), m.Authenticate())).Methods("DELETE")
	r.Handle("/v1/organizations/{organizationID}/idp/open-id", m.Chain(handler.GetIdentityProvider, m.Logger(), m.Authenticate())).Methods("GET")

	r.Handle("/v1/organizations/{orgID}/users/{userID}/consents", m.Chain(handler.GetConsents, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{orgID}/users/{userID}/consents/{consentID}", m.Chain(handler.GetConsentByID, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{orgID}/users/{userID}/consents/{consentID}/purposes/{purposeID}", m.Chain(handler.GetConsentPurposeByID, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{orgID}/purposes/{purposeID}/attributes/{attributeID}/consented/users", m.Chain(handler.GetAllUsersConsentedToAttribute, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{orgID}/users/{userID}/consents/{consentID}/purposes/{purposeID}/status", m.Chain(handler.GetPurposeAllConsentStatus, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{orgID}/users/{userID}/consents/{consentID}/purposes/{purposeID}/status", m.Chain(handler.UpdatePurposeAllConsentsv2, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{orgID}/purposes/{purposeID}/consented/users", m.Chain(handler.GetAllUsersConsentedToPurpose, m.Logger(), m.Authenticate())).Methods("GET")

	r.Handle("/v1/organizations/{orgID}/notify-data-breach", m.Chain(handler.NotifyDataBreach, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{orgID}/notify-events", m.Chain(handler.NotifyEvents, m.Logger(), m.Authenticate())).Methods("POST")

	r.Handle("/v1/organizations/{orgID}/data-requests", m.Chain(handler.GetDataRequests, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{orgID}/data-requests/{dataReqID}", m.Chain(handler.GetDataRequest, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{orgID}/data-requests/{dataReqID}", m.Chain(handler.UpdateDataRequests, m.Logger(), m.Authenticate())).Methods("PATCH")

	// Organisation webhooks related api(s)
	r.Handle("/v1/organizations/webhooks/payload/content-types", m.Chain(handler.GetWebhookPayloadContentTypes, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{orgID}/webhooks", m.Chain(handler.CreateWebhook, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{orgID}/webhooks", m.Chain(handler.GetAllWebhooks, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{orgID}/webhooks/{webhookID}", m.Chain(handler.GetWebhook, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{orgID}/webhooks/{webhookID}", m.Chain(handler.DeleteWebhook, m.Logger(), m.Authenticate())).Methods("DELETE")
	r.Handle("/v1/organizations/{orgID}/webhooks/{webhookID}", m.Chain(handler.UpdateWebhook, m.Logger(), m.Authenticate())).Methods("PUT")
	r.Handle("/v1/organizations/{orgID}/webhooks/{webhookID}/ping", m.Chain(handler.PingWebhook, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{orgID}/webhooks/{webhookID}/delivery", m.Chain(handler.GetRecentWebhookDeliveries, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{orgID}/webhooks/{webhookID}/delivery/{deliveryID}", m.Chain(handler.GetWebhookDeliveryByID, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{orgID}/webhooks/{webhookID}/delivery/{deliveryID}/redeliver", m.Chain(handler.ReDeliverWebhook, m.Logger(), m.Authenticate())).Methods("POST")

	// organization action logs
	r.Handle("/v1/organizations/{orgID}/logs", m.Chain(handler.GetOrgLogs, m.Logger(), m.Authenticate())).Methods("GET")

	//Login
	r.Handle("/v1/users/register", m.Chain(handler.RegisterUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1/users/login", m.Chain(handler.LoginUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1.1/users/login", m.Chain(handler.LoginUserV11, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1/users/validate/email", m.Chain(handler.ValidateUserEmail, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1/users/validate/phone", m.Chain(handler.ValidatePhoneNumber, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1/users/verify/phone", m.Chain(handler.VerifyPhoneNumber, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1/users/verify/otp", m.Chain(handler.VerifyOtp, m.LoggerNoAuth())).Methods("POST")

	// Admin login
	r.Handle("/v1/users/admin/login", m.Chain(handler.LoginAdminUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1/users/token", http.HandlerFunc(handler.GetToken)).Methods("POST")
	r.Handle("/v1/user/password/reset", m.Chain(handler.ResetPassword, m.Logger(), m.Authenticate())).Methods("PUT")
	r.Handle("/v1/user/password/forgot", m.Chain(handler.ForgotPassword, m.LoggerNoAuth())).Methods("PUT")
	r.Handle("/v1/users/logout", m.Chain(handler.LogoutUser, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/users/unregister", m.Chain(handler.UnregisterUser, m.Logger(), m.Authenticate())).Methods("POST")

	//user
	r.Handle("/v1/user", m.Chain(handler.GetCurrentUser, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/user", m.Chain(handler.UpdateCurrentUser, m.Logger(), m.Authenticate())).Methods("PATCH")
	r.Handle("/v1/user/register/ios", m.Chain(handler.UserClientRegister, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/user/register/android", m.Chain(handler.UserClientRegister, m.Logger(), m.Authenticate())).Methods("POST")

	r.Handle("/v1/user/apikey", m.Chain(handler.CreateAPIKey, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/user/apikey/revoke", m.Chain(handler.DeleteAPIKey, m.Logger(), m.Authenticate())).Methods("DELETE")
	r.Handle("/v1/user/apikey", m.Chain(handler.GetAPIKey, m.Logger(), m.Authenticate())).Methods("GET")

	//Consent History
	r.Handle("/v1/users/{userID}/consenthistory", m.Chain(handler.GetUserConsentHistory, m.Logger(), m.Authenticate())).Methods("GET")

}
