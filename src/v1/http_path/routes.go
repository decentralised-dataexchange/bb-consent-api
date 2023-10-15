package http_path

import (
	"net/http"

	m "github.com/bb-consent/api/src/middleware"
	v1Handler "github.com/bb-consent/api/src/v1/handler"
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

	r.Handle(AddOrganization, m.Chain(v1Handler.AddOrganization, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetOrganizationRoles, m.Chain(v1Handler.GetOrganizationRoles, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle(GetSubscribeMethods, m.Chain(v1Handler.GetSubscribeMethods, m.Logger(), m.Authenticate())).Methods("GET")

	r.Handle(GetDataRequestStatus, m.Chain(v1Handler.GetDataRequestStatus, m.Logger(), m.Authenticate())).Methods("GET")

	r.Handle(GetOrganizationTypes, m.Chain(v1Handler.GetOrganizationTypes, m.LoggerNoAuth())).Methods("GET")
	r.Handle(AddOrganizationType, m.Chain(v1Handler.AddOrganizationType, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(UpdateOrganizationType, m.Chain(v1Handler.UpdateOrganizationType, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PATCH")
	r.Handle(DeleteOrganizationType, m.Chain(v1Handler.DeleteOrganizationType, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(GetOrganizationTypeByID, m.Chain(v1Handler.GetOrganizationTypeByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(UpdateOrganizationTypeImage, m.Chain(v1Handler.UpdateOrganizationTypeImage, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetOrganizationTypeImage, m.Chain(v1Handler.GetOrganizationTypeImage, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	// Organization webhook event types
	r.Handle(GetWebhookEventTypes, m.Chain(v1Handler.GetWebhookEventTypes, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	r.Handle(GetOrganizationByID, m.Chain(v1Handler.GetOrganizationByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(UpdateOrganization, m.Chain(v1Handler.UpdateOrganization, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PATCH")
	r.Handle(UpdateOrganizationCoverImage, m.Chain(v1Handler.UpdateOrganizationCoverImage, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(UpdateOrganizationLogoImage, m.Chain(v1Handler.UpdateOrganizationLogoImage, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetOrganizationImage, m.Chain(v1Handler.GetOrganizationImage, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetOrganizationImageWeb, m.Chain(v1Handler.GetOrganizationImageWeb, m.LoggerNoAuth())).Methods("GET")

	r.Handle(UpdateOrgEula, m.Chain(v1Handler.UpdateOrgEula, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(DeleteOrgEula, m.Chain(v1Handler.DeleteOrgEula, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")

	r.Handle(AddOrgAdmin, m.Chain(v1Handler.AddOrgAdmin, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetOrgAdmins, m.Chain(v1Handler.GetOrgAdmins, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DeleteOrgAdmin, m.Chain(v1Handler.DeleteOrgAdmin, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")

	r.Handle(AddConsentPurposes, m.Chain(v1Handler.AddConsentPurposes, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetPurposes, m.Chain(v1Handler.GetPurposes, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DeleteConsentPurposeByID, m.Chain(v1Handler.DeleteConsentPurposeByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(UpdatePurposeByID, m.Chain(v1Handler.UpdatePurposeByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PUT")
	r.Handle(GetPurposeByID, m.Chain(v1Handler.GetPurposeByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	r.Handle(AddConsentTemplates, m.Chain(v1Handler.AddConsentTemplates, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetTemplates, m.Chain(v1Handler.GetTemplates, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DeleteConsentTemplateByID, m.Chain(v1Handler.DeleteConsentTemplateByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(DeleteConsentTemplatesByID, m.Chain(v1Handler.DeleteConsentTemplatesByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(GetTemplateByID, m.Chain(v1Handler.GetTemplateByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(UpdateTemplateByID, m.Chain(v1Handler.UpdateTemplateByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PUT")

	r.Handle(AddUserToOrganization, m.Chain(v1Handler.AddUserToOrganization, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(DeleteUserFromOrganization, m.Chain(v1Handler.DeleteUserFromOrganization, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(GetOrganizationUsers, m.Chain(v1Handler.GetOrganizationUsers, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetOrganizationUsersCount, m.Chain(v1Handler.GetOrganizationUsersCount, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	// Organization global policy configuration
	r.Handle(GetGlobalPolicyConfiguration, m.Chain(v1Handler.GetGlobalPolicyConfiguration, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(UpdateGlobalPolicyConfiguration, m.Chain(v1Handler.UpdateGlobalPolicyConfiguration, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")

	r.Handle(EnableOrganizationSubscription, m.Chain(v1Handler.EnableOrganizationSubscription, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(DisableOrganizationSubscription, m.Chain(v1Handler.DisableOrganizationSubscription, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetSubscribeMethod, m.Chain(v1Handler.GetSubscribeMethod, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(SetSubscribeMethod, m.Chain(v1Handler.SetSubscribeMethod, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetSubscribeKey, m.Chain(v1Handler.GetSubscribeKey, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(RenewSubscribeKey, m.Chain(v1Handler.RenewSubscribeKey, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetOrganizationSubscriptionStatus, m.Chain(v1Handler.GetOrganizationSubscriptionStatus, m.Logger(), m.Authenticate())).Methods("GET")

	// Organisation identity provider related API(s)
	r.Handle(AddIdentityProvider, m.Chain(v1Handler.AddIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(UpdateIdentityProvider, m.Chain(v1Handler.UpdateIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PUT")
	r.Handle(DeleteIdentityProvider, m.Chain(v1Handler.DeleteIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(GetIdentityProvider, m.Chain(v1Handler.GetIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	r.Handle(GetConsents, m.Chain(v1Handler.GetConsents, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetConsentByID, m.Chain(v1Handler.GetConsentByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetConsentPurposeByID, m.Chain(v1Handler.GetConsentPurposeByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetAllUsersConsentedToAttribute, m.Chain(v1Handler.GetAllUsersConsentedToAttribute, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetPurposeAllConsentStatus, m.Chain(v1Handler.GetPurposeAllConsentStatus, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(UpdatePurposeAllConsentsv2, m.Chain(v1Handler.UpdatePurposeAllConsentsv2, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(UpdatePurposeAttribute, m.Chain(v1Handler.UpdatePurposeAttribute, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PATCH")
	r.Handle(GetAllUsersConsentedToPurpose, m.Chain(v1Handler.GetAllUsersConsentedToPurpose, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	r.Handle(NotifyDataBreach, m.Chain(v1Handler.NotifyDataBreach, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle(NotifyEvents, m.Chain(v1Handler.NotifyEvents, m.Logger(), m.Authenticate())).Methods("POST")

	r.Handle(GetDataRequests, m.Chain(v1Handler.GetDataRequests, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetDataRequest, m.Chain(v1Handler.GetDataRequest, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(UpdateDataRequests, m.Chain(v1Handler.UpdateDataRequests, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PATCH")

	// Organisation webhooks related api(s)
	r.Handle(GetWebhookPayloadContentTypes, m.Chain(v1Handler.GetWebhookPayloadContentTypes, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(CreateWebhook, m.Chain(v1Handler.CreateWebhook, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetAllWebhooks, m.Chain(v1Handler.GetAllWebhooks, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetWebhook, m.Chain(v1Handler.GetWebhook, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DeleteWebhook, m.Chain(v1Handler.DeleteWebhook, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(UpdateWebhook, m.Chain(v1Handler.UpdateWebhook, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PUT")
	r.Handle(PingWebhook, m.Chain(v1Handler.PingWebhook, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetRecentWebhookDeliveries, m.Chain(v1Handler.GetRecentWebhookDeliveries, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(GetWebhookDeliveryByID, m.Chain(v1Handler.GetWebhookDeliveryByID, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(ReDeliverWebhook, m.Chain(v1Handler.ReDeliverWebhook, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")

	// organization action logs
	r.Handle(GetOrgLogs, m.Chain(v1Handler.GetOrgLogs, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	//Login
	r.Handle(RegisterUser, m.Chain(v1Handler.RegisterUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle(LoginUser, m.Chain(v1Handler.LoginUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle(LoginUserV11, m.Chain(v1Handler.LoginUserV11, m.LoggerNoAuth())).Methods("POST")
	r.Handle(ValidateUserEmail, m.Chain(v1Handler.ValidateUserEmail, m.LoggerNoAuth())).Methods("POST")
	r.Handle(ValidatePhoneNumber, m.Chain(v1Handler.ValidatePhoneNumber, m.LoggerNoAuth())).Methods("POST")
	r.Handle(VerifyPhoneNumber, m.Chain(v1Handler.VerifyPhoneNumber, m.LoggerNoAuth())).Methods("POST")
	r.Handle(VerifyOtp, m.Chain(v1Handler.VerifyOtp, m.LoggerNoAuth())).Methods("POST")

	// Admin login
	r.Handle(LoginAdminUser, m.Chain(v1Handler.LoginAdminUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle(GetToken, http.HandlerFunc(v1Handler.GetToken)).Methods("POST")
	r.Handle(ResetPassword, m.Chain(v1Handler.ResetPassword, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PUT")
	r.Handle(ForgotPassword, m.Chain(v1Handler.ForgotPassword, m.LoggerNoAuth())).Methods("PUT")
	r.Handle(LogoutUser, m.Chain(v1Handler.LogoutUser, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(UnregisterUser, m.Chain(v1Handler.UnregisterUser, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")

	//user
	r.Handle(GetCurrentUser, m.Chain(v1Handler.GetCurrentUser, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(UpdateCurrentUser, m.Chain(v1Handler.UpdateCurrentUser, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PATCH")
	r.Handle(UserClientRegisterIOS, m.Chain(v1Handler.UserClientRegister, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(UserClientRegisterAndroid, m.Chain(v1Handler.UserClientRegister, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")

	r.Handle(CreateAPIKey, m.Chain(v1Handler.CreateAPIKey, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(DeleteAPIKey, m.Chain(v1Handler.DeleteAPIKey, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(GetAPIKey, m.Chain(v1Handler.GetAPIKey, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	//Consent History
	r.Handle(GetUserConsentHistory, m.Chain(v1Handler.GetUserConsentHistory, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	r.Handle(GetMyOrgDataRequestStatus, m.Chain(v1Handler.GetMyOrgDataRequestStatus, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	r.Handle(GetDeleteMyData, m.Chain(v1Handler.GetDeleteMyData, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DeleteMyData, m.Chain(v1Handler.DeleteMyData, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetDeleteMyDataStatus, m.Chain(v1Handler.GetDeleteMyDataStatus, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DataDeleteCancelMyDataRequest, m.Chain(v1Handler.CancelMyDataRequest, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")

	r.Handle(GetDownloadMyData, m.Chain(v1Handler.GetDownloadMyData, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DownloadMyData, m.Chain(v1Handler.DownloadMyData, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetDownloadMyDataStatus, m.Chain(v1Handler.GetDownloadMyDataStatus, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(DataDownloadCancelMyDataRequest, m.Chain(v1Handler.CancelMyDataRequest, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(GetUserOrgsAndConsents, m.Chain(v1Handler.GetUserOrgsAndConsents, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

}
