package http_path

import (
	apiKeyHandler "github.com/bb-consent/api/src/v2/handler/apikey"
	auditHandler "github.com/bb-consent/api/src/v2/handler/audit"
	dataAgreementHandler "github.com/bb-consent/api/src/v2/handler/dataagreement"
	dataAttributeHandler "github.com/bb-consent/api/src/v2/handler/dataattribute"
	idpHandler "github.com/bb-consent/api/src/v2/handler/idp"
	individualHandler "github.com/bb-consent/api/src/v2/handler/individual"
	onboardHandler "github.com/bb-consent/api/src/v2/handler/onboard"
	policyHandler "github.com/bb-consent/api/src/v2/handler/policy"
	serviceHandler "github.com/bb-consent/api/src/v2/handler/service"
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

	r.Handle(ConfigCreateIndividualsInBulk, m.Chain(individualHandler.ConfigCreateIndividualsInBulk, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")

	// Service api(s)

	//  Data agreements
	r.Handle(ServiceReadDataAgreement, m.Chain(serviceHandler.ServiceReadDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ServiceListDataAgreements, m.Chain(serviceHandler.ServiceListDataAgreements, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Read an idp
	r.Handle(ServiceReadIdp, m.Chain(serviceHandler.ServiceReadIdp, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Policy
	r.Handle(ServiceReadPolicy, m.Chain(serviceHandler.ServiceReadPolicy, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Data attributes
	r.Handle(ServiceListDataAttributesForDataAgreement, m.Chain(serviceHandler.ServiceListDataAttributesForDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Verification mechanisms

	r.Handle(ServiceVerificationFetchAllDataAgreementRecords, m.Chain(serviceHandler.ServiceFetchDataAgreementRecords, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ServiceVerificationFetchDataAgreementRecord, m.Chain(serviceHandler.ServiceVerificationFetchDataAgreementRecord, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ServiceVerificationFetchDataAgreementRecords, m.Chain(serviceHandler.ServiceVerificationFetchDataAgreementRecords, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Recording consent
	r.Handle(ServiceCreateDraftConsentRecord, m.Chain(serviceHandler.ServiceCreateDraftConsentRecord, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(ServiceCreateDataAgreementRecord, m.Chain(serviceHandler.ServiceCreateDataAgreementRecord, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(ServiceUpdateDataAgreementRecord, m.Chain(serviceHandler.ServiceUpdateDataAgreementRecord, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(ServiceDeleteIndividualDataAgreementRecords, m.Chain(serviceHandler.ServiceDeleteIndividualDataAgreementRecords, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(ServiceCreatePairedDataAgreementRecord, m.Chain(serviceHandler.ServiceCreatePairedDataAgreementRecord, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(ServiceUpdateSignatureObject, m.Chain(serviceHandler.ServiceUpdateSignatureObject, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(ServiceCreateBlankSignature, m.Chain(serviceHandler.ServiceCreateBlankSignature, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")

	r.Handle(ServiceReadDataAgreementRecord, m.Chain(serviceHandler.ServiceReadDataAgreementRecord, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ServiceFetchIndividualDataAgreementRecords, m.Chain(serviceHandler.ServiceFetchIndividualDataAgreementRecords, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(ServiceFetchRecordsForDataAgreement, m.Chain(serviceHandler.ServiceFetchRecordsForDataAgreement, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	r.Handle(ServiceFetchRecordsHistory, m.Chain(serviceHandler.ServiceFetchRecordsHistory, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Audit api(s)

	r.Handle(AuditConsentRecordList, m.Chain(auditHandler.AuditConsentRecordList, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(AuditConsentRecordRead, m.Chain(auditHandler.AuditConsentRecordRead, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(AuditAgreementList, m.Chain(auditHandler.AuditAgreementList, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(AuditReadRecord, m.Chain(auditHandler.AuditReadRecord, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// organization action logs
	r.Handle(AuditGetOrgLogs, m.Chain(auditHandler.AuditGetOrgLogs, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Onboard api(s)

	r.Handle(LoginAdminUser, m.Chain(onboardHandler.LoginAdminUser, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	r.Handle(LoginUser, m.Chain(onboardHandler.LoginUser, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	r.Handle(OnboardResetPassword, m.Chain(onboardHandler.OnboardResetPassword, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")

	r.Handle(ValidateUserEmail, m.Chain(onboardHandler.ValidateUserEmail, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	r.Handle(ValidatePhoneNumber, m.Chain(onboardHandler.ValidatePhoneNumber, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	r.Handle(VerifyPhoneNumber, m.Chain(onboardHandler.VerifyPhoneNumber, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	r.Handle(VerifyOtp, m.Chain(onboardHandler.VerifyOtp, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")

	r.Handle(OnboardRefreshToken, m.Chain(onboardHandler.OnboardRefreshToken, m.AddContentType())).Methods("POST")
	r.Handle(ExchangeAuthorizationCode, m.Chain(onboardHandler.ExchangeAuthorizationCode, m.LoggerNoAuth())).Methods("POST")
	r.Handle(OnboardForgotPassword, m.Chain(onboardHandler.OnboardForgotPassword, m.LoggerNoAuth())).Methods("PUT")

	r.Handle(GetOrganizationByID, m.Chain(onboardHandler.OnboardReadOrganisation, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(UpdateOrganization, m.Chain(onboardHandler.UpdateOrganization, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(UpdateOrganizationCoverImage, m.Chain(onboardHandler.UpdateOrganizationCoverImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(UpdateOrganizationLogoImage, m.Chain(onboardHandler.UpdateOrganizationLogoImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("POST")
	r.Handle(GetOrganizationCoverImage, m.Chain(onboardHandler.GetOrganizationCoverImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(GetOrganizationLogoImage, m.Chain(onboardHandler.GetOrganizationLogoImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")

	r.Handle(OnboardReadOrganisationAdmin, m.Chain(onboardHandler.OnboardReadOrganisationAdmin, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(OnboardUpdateOrganisationAdmin, m.Chain(onboardHandler.OnboardUpdateOrganisationAdmin, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	r.Handle(OnboardReadOrganisationAdminAvatar, m.Chain(onboardHandler.OnboardReadOrganisationAdminAvatar, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("GET")
	r.Handle(OnboardUpdateOrganisationAdminAvatar, m.Chain(onboardHandler.OnboardUpdateOrganisationAdminAvatar, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.Authenticate(), m.AddContentType())).Methods("PUT")
}
