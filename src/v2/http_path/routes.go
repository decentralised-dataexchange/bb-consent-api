package http_path

import (
	m "github.com/bb-consent/api/src/middleware"
	v2Handler "github.com/bb-consent/api/src/v2/handler"
	dataAgreementHandler "github.com/bb-consent/api/src/v2/handler/dataagreement"
	policyHandler "github.com/bb-consent/api/src/v2/handler/policy"
	"github.com/casbin/casbin/v2"
	"github.com/gorilla/mux"
)

// SetRoutes sets the routes that the back end server serves
func SetRoutes(r *mux.Router, e *casbin.Enforcer) {
	// Policy
	r.Handle(ConfigReadPolicy, m.Chain(policyHandler.ConfigReadPolicy, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigCreatePolicy, m.Chain(policyHandler.ConfigCreatePolicy, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(ConfigUpdatePolicy, m.Chain(policyHandler.ConfigUpdatePolicy, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(ConfigListPolicyRevisions, m.Chain(policyHandler.ConfigListPolicyRevisions, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigDeletePolicy, m.Chain(policyHandler.ConfigDeletePolicy, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	r.Handle(ConfigListPolicies, m.Chain(policyHandler.ConfigListPolicies, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	r.Handle(ConfigReadDataAgreement, m.Chain(dataAgreementHandler.ConfigReadDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigCreateDataAgreement, m.Chain(dataAgreementHandler.ConfigCreateDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(ConfigUpdateDataAgreement, m.Chain(dataAgreementHandler.ConfigUpdateDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(ConfigListDataAgreementRevisions, m.Chain(dataAgreementHandler.ConfigListDataAgreementRevisions, m.Logger(), m.SetApplicationMode(), m.Authorize(e), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigDeleteDataAgreement, m.Chain(dataAgreementHandler.ConfigDeleteDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	r.Handle(ConfigListDataAgreements, m.Chain(dataAgreementHandler.ConfigListDataAgreements, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ReadDataAgreementRevision, m.Chain(v2Handler.ReadDataAgreementRevision, m.Logger(), m.SetApplicationMode(), m.Authorize(e), m.Authenticate(), m.AddContentType())).Methods("GET")

	r.Handle(GetDataAttributes, m.Chain(v2Handler.GetDataAttributes, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(AddDataAttribute, m.Chain(v2Handler.AddDataAttribute, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(UpdateDataAttributeById, m.Chain(v2Handler.UpdateDataAttributeById, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(DeleteDataAttributeById, m.Chain(v2Handler.DeleteDataAttributeById, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("DELETE")

	// Organisation webhooks related api(s)
	r.Handle(GetWebhookEventTypes, m.Chain(v2Handler.GetWebhookEventTypes, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(GetWebhookPayloadContentTypes, m.Chain(v2Handler.GetWebhookPayloadContentTypes, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(GetAllWebhooks, m.Chain(v2Handler.GetAllWebhooks, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(CreateWebhook, m.Chain(v2Handler.CreateWebhook, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(UpdateWebhook, m.Chain(v2Handler.UpdateWebhook, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(DeleteWebhook, m.Chain(v2Handler.DeleteWebhook, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	r.Handle(PingWebhook, m.Chain(v2Handler.PingWebhook, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(GetRecentWebhookDeliveries, m.Chain(v2Handler.GetRecentWebhookDeliveries, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(GetRecentWebhookDeliveryById, m.Chain(v2Handler.GetRecentWebhookDeliveryById, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(RedeliverWebhookPayloadByDeliveryID, m.Chain(v2Handler.RedeliverWebhookPayloadByDeliveryID, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")

	// Organisation identity provider related API(s)
	r.Handle(AddIdentityProvider, m.Chain(v2Handler.AddIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(UpdateIdentityProvider, m.Chain(v2Handler.UpdateIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(DeleteIdentityProvider, m.Chain(v2Handler.DeleteIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	r.Handle(GetIdentityProvider, m.Chain(v2Handler.GetIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Individual related api(s)
	r.Handle(GetOrganizationUsers, m.Chain(v2Handler.GetOrganizationUsers, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(RegisterUser, m.Chain(v2Handler.RegisterUser, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(GetUser, m.Chain(v2Handler.GetUser, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(DeleteUser, m.Chain(v2Handler.DeleteUser, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	r.Handle(UpdateUser, m.Chain(v2Handler.UpdateUser, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")

	// Api key related api(s)
	r.Handle(CreateAPIKey, m.Chain(v2Handler.CreateAPIKey, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(DeleteAPIKey, m.Chain(v2Handler.DeleteAPIKey, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	r.Handle(GetAPIKey, m.Chain(v2Handler.GetAPIKey, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Service api(s)

	//  Data agreements
	r.Handle(ServiceDataAgreementRead, m.Chain(dataAgreementHandler.ConfigReadDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Global policy configuration
	r.Handle(ServicePolicyRead, m.Chain(policyHandler.ConfigReadPolicy, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Data attributes
	r.Handle(ServiceGetDataAttributes, m.Chain(v2Handler.GetDataAttributes, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Verification mechanisms
	r.Handle(ServiceVerificationAgreementList, m.Chain(v2Handler.ServiceVerificationAgreementList, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ServiceVerificationAgreementConsentRecordRead, m.Chain(v2Handler.ServiceVerificationAgreementConsentRecordRead, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ServiceVerificationConsentRecordList, m.Chain(v2Handler.ServiceVerificationConsentRecordList, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Recording consent
	r.Handle(ServiceCreateIndividualConsentRecord, m.Chain(v2Handler.ServiceCreateIndividualConsentRecord, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(ServiceUpdateIndividualConsentRecord, m.Chain(v2Handler.ServiceCreateIndividualConsentRecord, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(ServiceListIndividualRecordList, m.Chain(v2Handler.ServiceListIndividualRecordList, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ServiceReadIndividualRecordRead, m.Chain(v2Handler.ServiceReadIndividualRecordRead, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Audit api(s)

	r.Handle(AuditConsentRecordList, m.Chain(v2Handler.AuditConsentRecordList, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(AuditConsentRecordRead, m.Chain(v2Handler.AuditConsentRecordRead, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(AuditAgreementList, m.Chain(v2Handler.AuditAgreementList, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(AuditReadRecord, m.Chain(v2Handler.AuditReadRecord, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// organization action logs
	r.Handle(GetOrgLogs, m.Chain(v2Handler.GetOrgLogs, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Onboard api(s)

	r.Handle(LoginAdminUser, m.Chain(v2Handler.LoginAdminUser, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	r.Handle(LoginUser, m.Chain(v2Handler.LoginUser, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")

	r.Handle(ValidateUserEmail, m.Chain(v2Handler.ValidateUserEmail, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	r.Handle(ValidatePhoneNumber, m.Chain(v2Handler.ValidatePhoneNumber, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	r.Handle(VerifyPhoneNumber, m.Chain(v2Handler.VerifyPhoneNumber, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	r.Handle(VerifyOtp, m.Chain(v2Handler.VerifyOtp, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")

	r.Handle(GetToken, m.Chain(v2Handler.GetToken)).Methods("POST")

	r.Handle(GetOrganizationByID, m.Chain(v2Handler.GetOrganizationByID, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(UpdateOrganization, m.Chain(v2Handler.UpdateOrganization, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(UpdateOrganizationCoverImage, m.Chain(v2Handler.UpdateOrganizationCoverImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(UpdateOrganizationLogoImage, m.Chain(v2Handler.UpdateOrganizationLogoImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(GetOrganizationCoverImage, m.Chain(v2Handler.GetOrganizationCoverImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(GetOrganizationLogoImage, m.Chain(v2Handler.GetOrganizationLogoImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
}
