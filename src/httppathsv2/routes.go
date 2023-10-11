package httppathsv2

import (
	"net/http"

	handler "github.com/bb-consent/api/src/handlerv2"
	m "github.com/bb-consent/api/src/middleware"
	"github.com/casbin/casbin/v2"
	"github.com/gorilla/mux"
)

// SetRoutes sets the routes that the back end server serves
func SetRoutes(r *mux.Router, e *casbin.Enforcer) {
	// Organization global policy configuration
	r.Handle(GetGlobalPolicyConfigurations, m.Chain(handler.GetGlobalPolicyConfiguration, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(UpdateGlobalPolicyConfigurations, m.Chain(handler.UpdateGlobalPolicyConfiguration, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("POST")
	r.Handle(UpdateGlobalPolicyConfigurationById, m.Chain(handler.UpdateGlobalPolicyConfigurationById, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("PUT")
	r.Handle(OrgListPolicyRevisions, m.Chain(handler.OrgListPolicyRevisions, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(OrgDeletePolicy, m.Chain(handler.OrgDeletePolicy, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("DELETE")
	r.Handle(OrgListPolicy, m.Chain(handler.OrgListPolicy, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")

	r.Handle(GetDataAgreementById, m.Chain(handler.GetDataAgreementById, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(AddDataAgreement, m.Chain(handler.AddDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("POST")
	r.Handle(UpdateDataAgreement, m.Chain(handler.UpdateDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("PUT")
	r.Handle(DeleteDataAgreement, m.Chain(handler.DeleteDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("DELETE")
	r.Handle(ListDataAgreements, m.Chain(handler.ListDataAgreements, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(ListDataAgreementRevisions, m.Chain(handler.ListDataAgreementRevisions, m.Logger(), m.SetApplicationMode(), m.Authorize(e), m.Authenticate())).Methods("GET")
	r.Handle(ReadDataAgreementRevision, m.Chain(handler.ReadDataAgreementRevision, m.Logger(), m.SetApplicationMode(), m.Authorize(e), m.Authenticate())).Methods("GET")

	r.Handle(GetDataAttributes, m.Chain(handler.GetDataAttributes, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(AddDataAttribute, m.Chain(handler.AddDataAttribute, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("POST")
	r.Handle(UpdateDataAttributeById, m.Chain(handler.UpdateDataAttributeById, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("PUT")
	r.Handle(DeleteDataAttributeById, m.Chain(handler.DeleteDataAttributeById, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("DELETE")

	// Organisation webhooks related api(s)
	r.Handle(GetWebhookEventTypes, m.Chain(handler.GetWebhookEventTypes, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(GetWebhookPayloadContentTypes, m.Chain(handler.GetWebhookPayloadContentTypes, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(GetAllWebhooks, m.Chain(handler.GetAllWebhooks, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(CreateWebhook, m.Chain(handler.CreateWebhook, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("POST")
	r.Handle(UpdateWebhook, m.Chain(handler.UpdateWebhook, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("PUT")
	r.Handle(DeleteWebhook, m.Chain(handler.DeleteWebhook, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("DELETE")
	r.Handle(PingWebhook, m.Chain(handler.PingWebhook, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("POST")
	r.Handle(GetRecentWebhookDeliveries, m.Chain(handler.GetRecentWebhookDeliveries, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(GetRecentWebhookDeliveryById, m.Chain(handler.GetRecentWebhookDeliveryById, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(RedeliverWebhookPayloadByDeliveryID, m.Chain(handler.RedeliverWebhookPayloadByDeliveryID, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("POST")

	// Organisation identity provider related API(s)
	r.Handle(AddIdentityProvider, m.Chain(handler.AddIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("POST")
	r.Handle(UpdateIdentityProvider, m.Chain(handler.UpdateIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("PUT")
	r.Handle(DeleteIdentityProvider, m.Chain(handler.DeleteIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("DELETE")
	r.Handle(GetIdentityProvider, m.Chain(handler.GetIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate())).Methods("GET")

	// Individual related api(s)
	r.Handle(GetOrganizationUsers, m.Chain(handler.GetOrganizationUsers, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(RegisterUser, m.Chain(handler.RegisterUser, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("POST")
	r.Handle(GetUser, m.Chain(handler.GetUser, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(DeleteUser, m.Chain(handler.DeleteUser, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("DELETE")
	r.Handle(UpdateUser, m.Chain(handler.UpdateUser, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("PUT")

	// Api key related api(s)
	r.Handle(CreateAPIKey, m.Chain(handler.CreateAPIKey, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("POST")
	r.Handle(DeleteAPIKey, m.Chain(handler.DeleteAPIKey, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("DELETE")
	r.Handle(GetAPIKey, m.Chain(handler.GetAPIKey, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")

	// Service api(s)

	//  Data agreements
	r.Handle(ServiceDataAgreementRead, m.Chain(handler.GetDataAgreementById, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")

	// Global policy configuration
	r.Handle(ServicePolicyRead, m.Chain(handler.GetGlobalPolicyConfiguration, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")

	// Data attributes
	r.Handle(ServiceGetDataAttributes, m.Chain(handler.GetDataAttributes, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")

	// Verification mechanisms
	r.Handle(ServiceVerificationAgreementList, m.Chain(handler.ServiceVerificationAgreementList, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(ServiceVerificationAgreementConsentRecordRead, m.Chain(handler.ServiceVerificationAgreementConsentRecordRead, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(ServiceVerificationConsentRecordList, m.Chain(handler.ServiceVerificationConsentRecordList, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")

	// Recording consent
	r.Handle(ServiceCreateIndividualConsentRecord, m.Chain(handler.ServiceCreateIndividualConsentRecord, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("POST")
	r.Handle(ServiceUpdateIndividualConsentRecord, m.Chain(handler.ServiceCreateIndividualConsentRecord, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("PUT")
	r.Handle(ServiceListIndividualRecordList, m.Chain(handler.ServiceListIndividualRecordList, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(ServiceReadIndividualRecordRead, m.Chain(handler.ServiceReadIndividualRecordRead, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")

	// Audit api(s)

	r.Handle(AuditConsentRecordList, m.Chain(handler.AuditConsentRecordList, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(AuditConsentRecordRead, m.Chain(handler.AuditConsentRecordRead, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(AuditAgreementList, m.Chain(handler.AuditAgreementList, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(AuditReadRecord, m.Chain(handler.AuditReadRecord, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")

	// organization action logs
	r.Handle(GetOrgLogs, m.Chain(handler.GetOrgLogs, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")

	// Onboard api(s)

	r.Handle(LoginAdminUser, m.Chain(handler.LoginAdminUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle(LoginUser, m.Chain(handler.LoginUser, m.LoggerNoAuth())).Methods("POST")

	r.Handle(ValidateUserEmail, m.Chain(handler.ValidateUserEmail, m.LoggerNoAuth())).Methods("POST")
	r.Handle(ValidatePhoneNumber, m.Chain(handler.ValidatePhoneNumber, m.LoggerNoAuth())).Methods("POST")
	r.Handle(VerifyPhoneNumber, m.Chain(handler.VerifyPhoneNumber, m.LoggerNoAuth())).Methods("POST")
	r.Handle(VerifyOtp, m.Chain(handler.VerifyOtp, m.LoggerNoAuth())).Methods("POST")

	r.Handle(GetToken, http.HandlerFunc(handler.GetToken)).Methods("POST")

	r.Handle(GetOrganizationByID, m.Chain(handler.GetOrganizationByID, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(UpdateOrganization, m.Chain(handler.UpdateOrganization, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("PUT")
	r.Handle(UpdateOrganizationCoverImage, m.Chain(handler.UpdateOrganizationCoverImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("POST")
	r.Handle(UpdateOrganizationLogoImage, m.Chain(handler.UpdateOrganizationLogoImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("POST")
	r.Handle(GetOrganizationCoverImage, m.Chain(handler.GetOrganizationImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
	r.Handle(GetOrganizationLogoImage, m.Chain(handler.GetOrganizationImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate())).Methods("GET")
}
