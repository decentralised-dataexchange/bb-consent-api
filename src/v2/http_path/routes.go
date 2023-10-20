package http_path

import (
	v2Handler "github.com/bb-consent/api/src/v2/handler"
	apiKeyHandler "github.com/bb-consent/api/src/v2/handler/apikey"
	dataAgreementHandler "github.com/bb-consent/api/src/v2/handler/dataagreement"
	dataAttributeHandler "github.com/bb-consent/api/src/v2/handler/dataattribute"
	idpHandler "github.com/bb-consent/api/src/v2/handler/idp"
	individualHandler "github.com/bb-consent/api/src/v2/handler/individual"
	onboardHandler "github.com/bb-consent/api/src/v2/handler/onboard"
	policyHandler "github.com/bb-consent/api/src/v2/handler/policy"
	webhookHandler "github.com/bb-consent/api/src/v2/handler/webhook"
	m "github.com/bb-consent/api/src/v2/middleware"
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

	// Data agreement
	r.Handle(ConfigReadDataAgreement, m.Chain(dataAgreementHandler.ConfigReadDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigCreateDataAgreement, m.Chain(dataAgreementHandler.ConfigCreateDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(ConfigUpdateDataAgreement, m.Chain(dataAgreementHandler.ConfigUpdateDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(ConfigListDataAgreementRevisions, m.Chain(dataAgreementHandler.ConfigListDataAgreementRevisions, m.Logger(), m.SetApplicationMode(), m.Authorize(e), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigDeleteDataAgreement, m.Chain(dataAgreementHandler.ConfigDeleteDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	r.Handle(ConfigListDataAgreements, m.Chain(dataAgreementHandler.ConfigListDataAgreements, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigListDataAttributesForDataAgreement, m.Chain(dataAgreementHandler.ConfigListDataAttributesForDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	r.Handle(ReadDataAgreementRevision, m.Chain(v2Handler.ReadDataAgreementRevision, m.Logger(), m.SetApplicationMode(), m.Authorize(e), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Data attribute
	r.Handle(ConfigReadDataAttribute, m.Chain(dataAttributeHandler.ConfigReadDataAttribute, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigCreateDataAttribute, m.Chain(dataAttributeHandler.ConfigCreateDataAttribute, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(ConfigUpdateDataAttribute, m.Chain(dataAttributeHandler.ConfigUpdateDataAttribute, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(ConfigListDataAttributeRevisions, m.Chain(dataAttributeHandler.ConfigListDataAttributeRevisions, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigDeleteDataAttribute, m.Chain(dataAttributeHandler.ConfigDeleteDataAttribute, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	r.Handle(ConfigListDataAttributes, m.Chain(dataAttributeHandler.ConfigListDataAttributes, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Organisation webhooks related api(s)
	r.Handle(ConfigReadWebhook, m.Chain(webhookHandler.ConfigReadWebhook, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigCreateWebhook, m.Chain(webhookHandler.ConfigCreateWebhook, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(ConfigUpdateWebhook, m.Chain(webhookHandler.ConfigUpdateWebhook, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(ConfigDeleteWebhook, m.Chain(webhookHandler.ConfigDeleteWebhook, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	r.Handle(ConfigListWebhooks, m.Chain(webhookHandler.ConfigListWebhooks, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigPingWebhook, m.Chain(webhookHandler.ConfigPingWebhook, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(ConfigListRecentWebhookDeliveries, m.Chain(webhookHandler.ConfigListRecentWebhookDeliveries, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigReadRecentWebhookDelivery, m.Chain(webhookHandler.ConfigReadRecentWebhookDelivery, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigRedeliverWebhookPayloadByDeliveryID, m.Chain(webhookHandler.ConfigRedeliverWebhookPayloadByDeliveryID, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(ConfigListWebhookEventTypes, m.Chain(webhookHandler.ConfigListWebhookEventTypes, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigListWebhookPayloadContentTypes, m.Chain(webhookHandler.ConfigListWebhookPayloadContentTypes, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Organisation identity provider related API(s)
	r.Handle(AddIdentityProvider, m.Chain(idpHandler.ConfigCreateIdp, m.Logger(), m.Authorize(e), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(UpdateIdentityProvider, m.Chain(idpHandler.UpdateIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(DeleteIdentityProvider, m.Chain(idpHandler.DeleteIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	r.Handle(GetIdentityProvider, m.Chain(idpHandler.GetIdentityProvider, m.Logger(), m.Authorize(e), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Individual related api(s)
	r.Handle(ConfigReadIndividual, m.Chain(individualHandler.ConfigReadIndividual, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ConfigCreateIndividual, m.Chain(individualHandler.ConfigCreateIndividual, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(ConfigUpdateIndividual, m.Chain(individualHandler.ConfigUpdateIndividual, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(ConfigDeleteIndividual, m.Chain(individualHandler.ConfigDeleteIndividual, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	r.Handle(ConfigListIndividuals, m.Chain(individualHandler.ConfigListIndividuals, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Api key related api(s)
	r.Handle(ConfigCreateApiKey, m.Chain(apiKeyHandler.ConfigCreateApiKey, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(ConfigDeleteApiKey, m.Chain(apiKeyHandler.ConfigDeleteApiKey, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	r.Handle(ConfigUpdateApiKey, m.Chain(apiKeyHandler.ConfigUpdateApiKey, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")

	// Service api(s)

	//  Data agreements
	r.Handle(ServiceDataAgreementRead, m.Chain(dataAgreementHandler.ConfigReadDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Global policy configuration
	r.Handle(ServicePolicyRead, m.Chain(policyHandler.ConfigReadPolicy, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Data attributes
	r.Handle(ServiceGetDataAttributes, m.Chain(dataAttributeHandler.ConfigListDataAttributes, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

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

	r.Handle(LoginAdminUser, m.Chain(onboardHandler.LoginAdminUser, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	r.Handle(LoginUser, m.Chain(onboardHandler.LoginUser, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")

	r.Handle(ValidateUserEmail, m.Chain(onboardHandler.ValidateUserEmail, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	r.Handle(ValidatePhoneNumber, m.Chain(onboardHandler.ValidatePhoneNumber, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	r.Handle(VerifyPhoneNumber, m.Chain(onboardHandler.VerifyPhoneNumber, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	r.Handle(VerifyOtp, m.Chain(onboardHandler.VerifyOtp, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")

	r.Handle(OnboardRefreshToken, m.Chain(onboardHandler.OnboardRefreshToken, m.AddContentType())).Methods("POST")
	r.Handle(ExchangeAuthorizationCode, m.Chain(onboardHandler.ExchangeAuthorizationCode, m.LoggerNoAuth())).Methods("POST")

	r.Handle(GetOrganizationByID, m.Chain(onboardHandler.OnboardReadOrganisation, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(UpdateOrganization, m.Chain(onboardHandler.UpdateOrganization, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(UpdateOrganizationCoverImage, m.Chain(onboardHandler.UpdateOrganizationCoverImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(UpdateOrganizationLogoImage, m.Chain(onboardHandler.UpdateOrganizationLogoImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(GetOrganizationCoverImage, m.Chain(onboardHandler.GetOrganizationCoverImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(GetOrganizationLogoImage, m.Chain(onboardHandler.GetOrganizationLogoImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
}
