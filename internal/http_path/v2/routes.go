package http_path

import (
	"net/http"

	auditHandler "github.com/bb-consent/api/internal/handler/v2/audit"
	apiKeyHandler "github.com/bb-consent/api/internal/handler/v2/config/apikey"
	dataAgreementHandler "github.com/bb-consent/api/internal/handler/v2/config/dataagreement"
	dataAttributeHandler "github.com/bb-consent/api/internal/handler/v2/config/dataattribute"
	idpHandler "github.com/bb-consent/api/internal/handler/v2/config/idp"
	configIndividualHandler "github.com/bb-consent/api/internal/handler/v2/config/individual"
	logHandler "github.com/bb-consent/api/internal/handler/v2/config/log"
	policyHandler "github.com/bb-consent/api/internal/handler/v2/config/policy"
	privacyDashboardHandler "github.com/bb-consent/api/internal/handler/v2/config/privacy_dashboard"
	webhookHandler "github.com/bb-consent/api/internal/handler/v2/config/webhook"
	onboardHandler "github.com/bb-consent/api/internal/handler/v2/onboard"
	serviceHandler "github.com/bb-consent/api/internal/handler/v2/service"
	serviceDataSharingHandler "github.com/bb-consent/api/internal/handler/v2/service/datasharing"
	serviceIndividualHandler "github.com/bb-consent/api/internal/handler/v2/service/individual"
	m "github.com/bb-consent/api/internal/middleware"
	"github.com/casbin/casbin/v2"
	"github.com/gorilla/mux"
)

// SetRoutes sets the routes that the back end server serves
func SetRoutes(r *mux.Router, e *casbin.Enforcer, testMode bool) {
	var wrapper = func(path string, handler http.Handler) *mux.Route {
		if testMode {
			path = path + "/"
		}
		return r.Handle(path, handler)
	}

	// Policy
	wrapper(ConfigReadPolicy, m.Chain(policyHandler.ConfigReadPolicy, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ConfigCreatePolicy, m.Chain(policyHandler.ConfigCreatePolicy, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")
	wrapper(ConfigUpdatePolicy, m.Chain(policyHandler.ConfigUpdatePolicy, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	wrapper(ConfigListPolicyRevisions, m.Chain(policyHandler.ConfigListPolicyRevisions, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ConfigDeletePolicy, m.Chain(policyHandler.ConfigDeletePolicy, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	wrapper(ConfigListPolicies, m.Chain(policyHandler.ConfigListPolicies, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Data agreement
	wrapper(ConfigReadDataAgreement, m.Chain(dataAgreementHandler.ConfigReadDataAgreement, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ConfigCreateDataAgreement, m.Chain(dataAgreementHandler.ConfigCreateDataAgreement, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")
	wrapper(ConfigUpdateDataAgreement, m.Chain(dataAgreementHandler.ConfigUpdateDataAgreement, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	wrapper(ConfigListDataAgreementRevisions, m.Chain(dataAgreementHandler.ConfigListDataAgreementRevisions, m.Logger(), m.LogApiCalls(), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authorize(e), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ConfigDeleteDataAgreement, m.Chain(dataAgreementHandler.ConfigDeleteDataAgreement, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	wrapper(ConfigListDataAgreements, m.Chain(dataAgreementHandler.ConfigListDataAgreements, m.Logger(), m.LogApiCalls(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ConfigListDataAttributesForDataAgreement, m.Chain(dataAgreementHandler.ConfigListDataAttributesForDataAgreement, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Data attribute
	wrapper(ConfigUpdateDataAttribute, m.Chain(dataAttributeHandler.ConfigUpdateDataAttribute, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	wrapper(ConfigListDataAttributes, m.Chain(dataAttributeHandler.ConfigListDataAttributes, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Organisation webhooks related api(s)
	wrapper(ConfigReadWebhook, m.Chain(webhookHandler.ConfigReadWebhook, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ConfigCreateWebhook, m.Chain(webhookHandler.ConfigCreateWebhook, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")
	wrapper(ConfigUpdateWebhook, m.Chain(webhookHandler.ConfigUpdateWebhook, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	wrapper(ConfigDeleteWebhook, m.Chain(webhookHandler.ConfigDeleteWebhook, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	wrapper(ConfigListWebhooks, m.Chain(webhookHandler.ConfigListWebhooks, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ConfigPingWebhook, m.Chain(webhookHandler.ConfigPingWebhook, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")
	wrapper(ConfigListRecentWebhookDeliveries, m.Chain(webhookHandler.ConfigListRecentWebhookDeliveries, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ConfigReadRecentWebhookDelivery, m.Chain(webhookHandler.ConfigReadRecentWebhookDelivery, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ConfigRedeliverWebhookPayloadByDeliveryID, m.Chain(webhookHandler.ConfigRedeliverWebhookPayloadByDeliveryID, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")
	wrapper(ConfigListWebhookEventTypes, m.Chain(webhookHandler.ConfigListWebhookEventTypes, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ConfigListWebhookPayloadContentTypes, m.Chain(webhookHandler.ConfigListWebhookPayloadContentTypes, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Organisation identity provider related API(s)
	wrapper(AddIdentityProvider, m.Chain(idpHandler.ConfigCreateIdp, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")
	wrapper(UpdateIdentityProvider, m.Chain(idpHandler.UpdateIdentityProvider, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	wrapper(DeleteIdentityProvider, m.Chain(idpHandler.DeleteIdentityProvider, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	wrapper(GetIdentityProvider, m.Chain(idpHandler.GetIdentityProvider, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ConfigListIdentityProviders, m.Chain(idpHandler.ConfigListIdps, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Individual related api(s)
	wrapper(ConfigReadIndividual, m.Chain(configIndividualHandler.ConfigReadIndividual, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ConfigCreateIndividual, m.Chain(configIndividualHandler.ConfigCreateIndividual, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")
	wrapper(ConfigUpdateIndividual, m.Chain(configIndividualHandler.ConfigUpdateIndividual, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	wrapper(ConfigListIndividuals, m.Chain(configIndividualHandler.ConfigListIndividuals, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Api key related api(s)
	wrapper(ConfigCreateApiKey, m.Chain(apiKeyHandler.ConfigCreateApiKey, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")
	wrapper(ConfigDeleteApiKey, m.Chain(apiKeyHandler.ConfigDeleteApiKey, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	wrapper(ConfigUpdateApiKey, m.Chain(apiKeyHandler.ConfigUpdateApiKey, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	wrapper(ConfigListApiKey, m.Chain(apiKeyHandler.ConfigListApiKey, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")

	wrapper(ConfigCreateIndividualsInBulk, m.Chain(configIndividualHandler.ConfigCreateIndividualsInBulk, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")

	wrapper(ConfigReadPrivacyDashboard, m.Chain(privacyDashboardHandler.ConfigReadPrivacyDashboard, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Purge logs
	wrapper(ConfigPurgeOrgLogs, m.Chain(logHandler.ConfigPurgeOrgLogs, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("DELETE")

	// Service api(s)

	//  Data agreements
	wrapper(ServiceReadDataAgreement, m.Chain(serviceHandler.ServiceReadDataAgreement, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKey(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ServiceListDataAgreements, m.Chain(serviceHandler.ServiceListDataAgreements, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKey(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Read an idp
	wrapper(ServiceReadIdp, m.Chain(serviceHandler.ServiceReadIdp, m.LoggerNoAuth(), m.SetApplicationMode(), m.AddContentType())).Methods("GET")

	// Policy
	wrapper(ServiceReadPolicy, m.Chain(serviceHandler.ServiceReadPolicy, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKey(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Data attributes
	wrapper(ServiceListDataAttributesForDataAgreement, m.Chain(serviceHandler.ServiceListDataAttributesForDataAgreement, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKey(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Verification mechanisms

	wrapper(ServiceVerificationListDataAgreements, m.Chain(serviceHandler.ServiceVerificationListDataAgreements, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKey(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ServiceVerificationFetchDataAgreementRecord, m.Chain(serviceHandler.ServiceVerificationFetchDataAgreementRecord, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKey(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ServiceVerificationFetchDataAgreementRecords, m.Chain(serviceHandler.ServiceVerificationFetchDataAgreementRecords, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKey(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Recording consent
	wrapper(ServiceCreateDraftConsentRecord, m.Chain(serviceHandler.ServiceCreateDraftConsentRecord, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")
	wrapper(ServiceCreateDataAgreementRecord, m.Chain(serviceHandler.ServiceCreateDataAgreementRecord, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")
	wrapper(ServiceUpdateDataAgreementRecord, m.Chain(serviceHandler.ServiceUpdateDataAgreementRecord, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	wrapper(ServiceDeleteIndividualDataAgreementRecords, m.Chain(serviceHandler.ServiceDeleteIndividualDataAgreementRecords, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("DELETE")
	wrapper(ServiceCreatePairedDataAgreementRecord, m.Chain(serviceHandler.ServiceCreatePairedDataAgreementRecord, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")
	wrapper(ServiceUpdateSignatureObject, m.Chain(serviceHandler.ServiceUpdateSignatureObject, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	wrapper(ServiceCreateBlankSignature, m.Chain(serviceHandler.ServiceCreateBlankSignature, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")

	wrapper(ServiceReadDataAgreementRecord, m.Chain(serviceHandler.ServiceReadDataAgreementRecord, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ServiceFetchIndividualDataAgreementRecords, m.Chain(serviceHandler.ServiceFetchIndividualDataAgreementRecords, m.ValidateIndividualId(), m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ServiceFetchRecordsForDataAgreement, m.Chain(serviceHandler.ServiceFetchRecordsForDataAgreement, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")

	wrapper(ServiceFetchRecordsHistory, m.Chain(serviceHandler.ServiceFetchRecordsHistory, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")

	wrapper(ServiceReadOrganisation, m.Chain(serviceHandler.ServiceReadOrganisation, m.LoggerNoAuth(), m.SetApplicationMode(), m.AddContentType())).Methods("GET")
	wrapper(ServiceReadOrganisationLogoImage, m.Chain(serviceHandler.ServiceReadOrganisationLogoImage, m.LoggerNoAuth(), m.SetApplicationMode(), m.AddContentType())).Methods("GET")
	wrapper(ServiceReadOrganisationCoverImage, m.Chain(serviceHandler.ServiceReadOrganisationCoverImage, m.LoggerNoAuth(), m.SetApplicationMode(), m.AddContentType())).Methods("GET")
	wrapper(ServiceReadOrganisationImage, m.Chain(serviceHandler.ServiceReadOrganisationImage, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKey(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Individual related api(s)
	wrapper(ServiceReadIndividual, m.Chain(serviceIndividualHandler.ServiceReadIndividual, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKey(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(ServiceCreateIndividual, m.Chain(serviceIndividualHandler.ServiceCreateIndividual, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKey(), m.Authenticate(), m.AddContentType())).Methods("POST")
	wrapper(ServiceUpdateIndividual, m.Chain(serviceIndividualHandler.ServiceUpdateIndividual, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKey(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	wrapper(ServiceListIndividuals, m.Chain(serviceIndividualHandler.ServiceListIndividuals, m.Logger(), m.ValidateIndividualId(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKey(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Audit api(s)

	wrapper(AuditListDataAgreementRecords, m.Chain(auditHandler.AuditListDataAgreementRecords, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(AuditDataAgreementRecordRead, m.Chain(auditHandler.AuditDataAgreementRecordRead, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(AuditListDataAgreements, m.Chain(auditHandler.AuditListDataAgreements, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(AuditReadDataAgreement, m.Chain(auditHandler.AuditReadDataAgreement, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// organization action logs
	wrapper(AuditGetOrgLogs, m.Chain(auditHandler.AuditGetOrgLogs, m.Logger(), m.LogApiCalls(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")

	// Onboard api(s)

	wrapper(LoginAdminUser, m.Chain(onboardHandler.LoginAdminUser, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	wrapper(LoginUser, m.Chain(onboardHandler.LoginUser, m.LoggerNoAuth(), m.AddContentType())).Methods("POST")
	wrapper(OnboardResetPassword, m.Chain(onboardHandler.OnboardResetPassword, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	wrapper(OnboardLogoutUser, m.Chain(onboardHandler.OnboardLogoutUser, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")

	wrapper(OnboardRefreshToken, m.Chain(onboardHandler.OnboardRefreshToken, m.AddContentType())).Methods("POST")
	wrapper(ExchangeAuthorizationCode, m.Chain(onboardHandler.ExchangeAuthorizationCode, m.LoggerNoAuth(), m.SetApplicationMode())).Methods("POST")
	wrapper(OnboardForgotPassword, m.Chain(onboardHandler.OnboardForgotPassword, m.LoggerNoAuth(), m.SetApplicationMode())).Methods("PUT")

	wrapper(GetOrganizationByID, m.Chain(onboardHandler.OnboardReadOrganisation, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(UpdateOrganization, m.Chain(onboardHandler.UpdateOrganization, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	wrapper(UpdateOrganizationCoverImage, m.Chain(onboardHandler.UpdateOrganizationCoverImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")
	wrapper(UpdateOrganizationLogoImage, m.Chain(onboardHandler.UpdateOrganizationLogoImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("POST")
	wrapper(GetOrganizationCoverImage, m.Chain(onboardHandler.GetOrganizationCoverImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(GetOrganizationLogoImage, m.Chain(onboardHandler.GetOrganizationLogoImage, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")

	wrapper(OnboardReadOrganisationAdmin, m.Chain(onboardHandler.OnboardReadOrganisationAdmin, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(OnboardUpdateOrganisationAdmin, m.Chain(onboardHandler.OnboardUpdateOrganisationAdmin, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("PUT")
	wrapper(OnboardReadOrganisationAdminAvatar, m.Chain(onboardHandler.OnboardReadOrganisationAdminAvatar, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")
	wrapper(OnboardUpdateOrganisationAdminAvatar, m.Chain(onboardHandler.OnboardUpdateOrganisationAdminAvatar, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("PUT")

	wrapper(OnboardReadStatus, m.Chain(onboardHandler.OnboardReadStatus, m.Logger(), m.Authorize(e), m.SetApplicationMode(), m.ValidateAPIKeyAndIndividualId(), m.Authenticate(), m.AddContentType())).Methods("GET")

	wrapper(ServiceShowDataSharingUi, m.Chain(serviceDataSharingHandler.ServiceShowDataSharingUiHandler, m.LoggerNoAuth())).Methods("GET")
}
