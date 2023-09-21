package main

import (
	"net/http"

	"github.com/bb-consent/api/src/handler"
	m "github.com/bb-consent/api/src/middleware"
	"github.com/casbin/casbin/v2"
	"github.com/gorilla/mux"
)

// Root access return 200 OK for health check when the api
// is deployed in K8s with ingress controller.
func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

// SetRoutes sets the routes that the back end server serves
func SetRoutes(r *mux.Router, e *casbin.Enforcer) {
	r.HandleFunc("/", healthz).Methods("GET")

	r.Handle(AddOrganization, m.Chain(handler.AddOrganization, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetOrganizationRoles, m.Chain(handler.GetOrganizationRoles, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle(GetSubscribeMethods, m.Chain(handler.GetSubscribeMethods, m.Logger(), m.Authenticate())).Methods("GET")

	r.Handle(GetDataRequestStatus, m.Chain(handler.GetDataRequestStatus, m.Logger(), m.Authenticate())).Methods("GET")

	r.Handle(GetOrganizationTypes, m.Chain(handler.GetOrganizationTypes, m.LoggerNoAuth())).Methods("GET")
	r.Handle(AddOrganizationType, m.Chain(handler.AddOrganizationType, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(UpdateOrganizationType, m.Chain(handler.UpdateOrganizationType, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PATCH")
	r.Handle(DeleteOrganizationType, m.Chain(handler.DeleteOrganizationType, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(GetOrganizationTypeByID, m.Chain(handler.GetOrganizationTypeByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(UpdateOrganizationTypeImage, m.Chain(handler.UpdateOrganizationTypeImage, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetOrganizationTypeImage, m.Chain(handler.GetOrganizationTypeImage, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	// Organization webhook event types
	r.Handle(GetWebhookEventTypes, m.Chain(handler.GetWebhookEventTypes, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	r.Handle(GetOrganizationByID, m.Chain(handler.GetOrganizationByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(UpdateOrganization, m.Chain(handler.UpdateOrganization, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PATCH")
	r.Handle(UpdateOrganizationCoverImage, m.Chain(handler.UpdateOrganizationCoverImage, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(UpdateOrganizationLogoImage, m.Chain(handler.UpdateOrganizationLogoImage, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetOrganizationImage, m.Chain(handler.GetOrganizationImage, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetOrganizationImageWeb, m.Chain(handler.GetOrganizationImageWeb, m.LoggerNoAuth())).Methods("GET")

	r.Handle(UpdateOrgEula, m.Chain(handler.UpdateOrgEula, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(DeleteOrgEula, m.Chain(handler.DeleteOrgEula, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")

	r.Handle(AddOrgAdmin, m.Chain(handler.AddOrgAdmin, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetOrgAdmins, m.Chain(handler.GetOrgAdmins, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DeleteOrgAdmin, m.Chain(handler.DeleteOrgAdmin, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")

	r.Handle(AddConsentPurposes, m.Chain(handler.AddConsentPurposes, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetPurposes, m.Chain(handler.GetPurposes, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DeleteConsentPurposeByID, m.Chain(handler.DeleteConsentPurposeByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(UpdatePurposeByID, m.Chain(handler.UpdatePurposeByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PUT")
	r.Handle(GetPurposeByID, m.Chain(handler.GetPurposeByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	r.Handle(AddConsentTemplates, m.Chain(handler.AddConsentTemplates, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetTemplates, m.Chain(handler.GetTemplates, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DeleteConsentTemplateByID, m.Chain(handler.DeleteConsentTemplateByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(DeleteConsentTemplatesByID, m.Chain(handler.DeleteConsentTemplatesByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(GetTemplateByID, m.Chain(handler.GetTemplateByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(UpdateTemplateByID, m.Chain(handler.UpdateTemplateByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PUT")

	r.Handle(AddUserToOrganization, m.Chain(handler.AddUserToOrganization, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(DeleteUserFromOrganization, m.Chain(handler.DeleteUserFromOrganization, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(GetOrganizationUsers, m.Chain(handler.GetOrganizationUsers, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetOrganizationUsersCount, m.Chain(handler.GetOrganizationUsersCount, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	// Organization global policy configuration
	r.Handle(GetGlobalPolicyConfiguration, m.Chain(handler.GetGlobalPolicyConfiguration, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(UpdateGlobalPolicyConfiguration, m.Chain(handler.UpdateGlobalPolicyConfiguration, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")

	r.Handle(EnableOrganizationSubscription, m.Chain(handler.EnableOrganizationSubscription, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(DisableOrganizationSubscription, m.Chain(handler.DisableOrganizationSubscription, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetSubscribeMethod, m.Chain(handler.GetSubscribeMethod, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(SetSubscribeMethod, m.Chain(handler.SetSubscribeMethod, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetSubscribeKey, m.Chain(handler.GetSubscribeKey, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(RenewSubscribeKey, m.Chain(handler.RenewSubscribeKey, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetOrganizationSubscriptionStatus, m.Chain(handler.GetOrganizationSubscriptionStatus, m.Logger(), m.Authenticate())).Methods("GET")

	// Organisation identity provider related API(s)
	r.Handle(AddIdentityProvider, m.Chain(handler.AddIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(UpdateIdentityProvider, m.Chain(handler.UpdateIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PUT")
	r.Handle(DeleteIdentityProvider, m.Chain(handler.DeleteIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(GetIdentityProvider, m.Chain(handler.GetIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	r.Handle(GetConsents, m.Chain(handler.GetConsents, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetConsentByID, m.Chain(handler.GetConsentByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetConsentPurposeByID, m.Chain(handler.GetConsentPurposeByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetAllUsersConsentedToAttribute, m.Chain(handler.GetAllUsersConsentedToAttribute, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetPurposeAllConsentStatus, m.Chain(handler.GetPurposeAllConsentStatus, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(UpdatePurposeAllConsentsv2, m.Chain(handler.UpdatePurposeAllConsentsv2, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(UpdatePurposeAttribute, m.Chain(handler.UpdatePurposeAttribute, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PATCH")
	r.Handle(GetAllUsersConsentedToPurpose, m.Chain(handler.GetAllUsersConsentedToPurpose, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	r.Handle(NotifyDataBreach, m.Chain(handler.NotifyDataBreach, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle(NotifyEvents, m.Chain(handler.NotifyEvents, m.Logger(), m.Authenticate())).Methods("POST")

	r.Handle(GetDataRequests, m.Chain(handler.GetDataRequests, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetDataRequest, m.Chain(handler.GetDataRequest, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(UpdateDataRequests, m.Chain(handler.UpdateDataRequests, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PATCH")

	// Organisation webhooks related api(s)
	r.Handle(GetWebhookPayloadContentTypes, m.Chain(handler.GetWebhookPayloadContentTypes, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(CreateWebhook, m.Chain(handler.CreateWebhook, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetAllWebhooks, m.Chain(handler.GetAllWebhooks, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetWebhook, m.Chain(handler.GetWebhook, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DeleteWebhook, m.Chain(handler.DeleteWebhook, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(UpdateWebhook, m.Chain(handler.UpdateWebhook, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PUT")
	r.Handle(PingWebhook, m.Chain(handler.PingWebhook, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetRecentWebhookDeliveries, m.Chain(handler.GetRecentWebhookDeliveries, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetWebhookDeliveryByID, m.Chain(handler.GetWebhookDeliveryByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(ReDeliverWebhook, m.Chain(handler.ReDeliverWebhook, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")

	// organization action logs
	r.Handle(GetOrgLogs, m.Chain(handler.GetOrgLogs, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	//Login
	r.Handle(RegisterUser, m.Chain(handler.RegisterUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle(LoginUser, m.Chain(handler.LoginUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle(LoginUserV11, m.Chain(handler.LoginUserV11, m.LoggerNoAuth())).Methods("POST")
	r.Handle(ValidateUserEmail, m.Chain(handler.ValidateUserEmail, m.LoggerNoAuth())).Methods("POST")
	r.Handle(ValidatePhoneNumber, m.Chain(handler.ValidatePhoneNumber, m.LoggerNoAuth())).Methods("POST")
	r.Handle(VerifyPhoneNumber, m.Chain(handler.VerifyPhoneNumber, m.LoggerNoAuth())).Methods("POST")
	r.Handle(VerifyOtp, m.Chain(handler.VerifyOtp, m.LoggerNoAuth())).Methods("POST")

	// Admin login
	r.Handle(LoginAdminUser, m.Chain(handler.LoginAdminUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle(GetToken, http.HandlerFunc(handler.GetToken)).Methods("POST")
	r.Handle(ResetPassword, m.Chain(handler.ResetPassword, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PUT")
	r.Handle(ForgotPassword, m.Chain(handler.ForgotPassword, m.LoggerNoAuth())).Methods("PUT")
	r.Handle(LogoutUser, m.Chain(handler.LogoutUser, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(UnregisterUser, m.Chain(handler.UnregisterUser, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")

	//user
	r.Handle(GetCurrentUser, m.Chain(handler.GetCurrentUser, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(UpdateCurrentUser, m.Chain(handler.UpdateCurrentUser, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PATCH")
	r.Handle(UserClientRegisterIOS, m.Chain(handler.UserClientRegister, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(UserClientRegisterAndroid, m.Chain(handler.UserClientRegister, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")

	r.Handle(CreateAPIKey, m.Chain(handler.CreateAPIKey, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(DeleteAPIKey, m.Chain(handler.DeleteAPIKey, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(GetAPIKey, m.Chain(handler.GetAPIKey, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	//Consent History
	r.Handle(GetUserConsentHistory, m.Chain(handler.GetUserConsentHistory, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	r.Handle(GetMyOrgDataRequestStatus, m.Chain(handler.GetMyOrgDataRequestStatus, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	r.Handle(GetDeleteMyData, m.Chain(handler.GetDeleteMyData, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DeleteMyData, m.Chain(handler.DeleteMyData, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetDeleteMyDataStatus, m.Chain(handler.GetDeleteMyDataStatus, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DataDeleteCancelMyDataRequest, m.Chain(handler.CancelMyDataRequest, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")

	r.Handle(GetDownloadMyData, m.Chain(handler.GetDownloadMyData, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DownloadMyData, m.Chain(handler.DownloadMyData, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetDownloadMyDataStatus, m.Chain(handler.GetDownloadMyDataStatus, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DataDownloadCancelMyDataRequest, m.Chain(handler.CancelMyDataRequest, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetUserOrgsAndConsents, m.Chain(handler.GetUserOrgsAndConsents, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

}
