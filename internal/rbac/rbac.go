package rbac

import "github.com/casbin/casbin/v2/model"

// RBAC User Roles
const (
	ROLE_USER  string = "user"
	ROLE_ADMIN string = "organisation_admin"
)

// GetRbacPolicies
func GetRbacPolicies(testMode bool) [][]string {

	policies := [][]string{
		{"organisation_admin", "/config/policy", "POST"},
		{"organisation_admin", "/config/policy/{policyId}", "(GET)|(PUT)|(DELETE)"},
		{"organisation_admin", "/config/policy/{policyId}/revisions", "GET"},
		{"organisation_admin", "/config/policies", "GET"},
		{"organisation_admin", "/config/data-agreement/{dataAgreementId}", "(GET)|(PUT)|(DELETE)"},
		{"organisation_admin", "/config/data-agreement", "POST"},
		{"organisation_admin", "/config/data-agreements", "GET"},
		{"organisation_admin", "/config/data-agreement/{dataAgreementId}/revisions", "GET"},
		{"organisation_admin", "/config/data-agreement/{dataAgreementId}/revision/{revisionId}", "GET"},
		{"organisation_admin", "/config/data-agreement/{dataAgreementId}/data-attributes", "GET"},
		{"organisation_admin", "/config/data-agreements/data-attribute", "POST"},
		{"organisation_admin", "/config/data-agreements/data-attribute/{dataAttributeId}", "(GET)|(PUT)|(DELETE)"},
		{"organisation_admin", "/config/data-agreements/data-attribute/{dataAttributeId}/revisions", "GET"},
		{"organisation_admin", "/config/data-agreements/data-attributes", "GET"},
		{"organisation_admin", "/config/webhooks/event-types", "GET"},
		{"organisation_admin", "/config/webhooks/payload/content-types", "GET"},
		{"organisation_admin", "/config/webhooks", "GET"},
		{"organisation_admin", "/config/webhook", "POST"},
		{"organisation_admin", "/config/webhook/{webhookId}", "(GET)|(PUT)|(DELETE)"},
		{"organisation_admin", "/config/webhook/{webhookId}/ping", "POST"},
		{"organisation_admin", "/config/webhooks/{webhookId}/deliveries", "GET"},
		{"organisation_admin", "/config/webhooks/{webhookId}/delivery/{deliveryId}", "GET"},
		{"organisation_admin", "/config/webhooks/{webhookId}/delivery/{deliveryId}/redeliver", "POST"},
		{"organisation_admin", "/config/idp/open-id", "POST"},
		{"organisation_admin", "/config/idp/open-ids", "GET"},
		{"organisation_admin", "/config/idp/open-id/{idpId}", "(GET)|(PUT)|(DELETE)"},
		{"organisation_admin", "/config/individuals", "GET"},
		{"organisation_admin", "/config/individual", "POST"},
		{"organisation_admin", "/config/individual/{individualId}", "(GET)|(PUT)"},
		{"organisation_admin", "/config/admin/apikey", "POST"},
		{"organisation_admin", "/config/admin/apikey/{apiKeyId}", "(PUT)|(DELETE)"},
		{"organisation_admin", "/config/admin/apikeys", "GET"},
		{"user", "/service/data-agreements", "GET"},
		{"user", "/service/data-agreement/{dataAgreementId}", "GET"},
		{"user", "/service/data-agreement/{dataAgreementId}/data-attributes", "GET"},
		{"user", "/service/policy/{policyId}", "GET"},
		{"user", "/service/verification/data-agreements", "GET"},
		{"user", "/service/verification/consent-record/{consentRecordId}", "GET"},
		{"user", "/service/verification/consent-records", "GET"},
		{"user", "/service/individual/record/consent-record/draft", "POST"},
		{"user", "/service/individual/record/data-agreement/{dataAgreementId}", "(GET)|(POST)"},
		{"user", "/service/individual/record/consent-record/{consentRecordId}", "PUT"},
		{"user", "/service/individual/record/consent-record", "(GET)|(POST)"},
		{"user", "/service/individual/record/consent-record/{consentRecordId}/signature", "(POST)|(PUT)"},
		{"user", "/service/individual/record/data-agreement/{dataAgreementId}/all", "GET"},
		{"organisation_admin", "/audit/consent-records", "GET"},
		{"organisation_admin", "/audit/consent-record/{consentRecordId}", "GET"},
		{"organisation_admin", "/audit/data-agreements", "GET"},
		{"organisation_admin", "/audit/data-agreement/{dataAgreementId}", "GET"},
		{"organisation_admin", "/audit/admin/logs", "GET"},
		{"organisation_admin", "/onboard/organisation", "(GET)|(PUT)"},
		{"organisation_admin", "/onboard/organisation/coverimage", "(GET)|(POST)"},
		{"organisation_admin", "/onboard/organisation/logoimage", "(GET)|(POST)"},
		{"user", "/onboard/organisation", "GET"},
		{"user", "/onboard/organisation/coverimage", "GET"},
		{"user", "/onboard/organisation/logoimage", "GET"},
		{"organisation_admin", "/onboard/password/reset", "PUT"},
		{"organisation_admin", "/onboard/admin", "(GET)|(PUT)"},
		{"organisation_admin", "/onboard/admin/avatarimage", "(GET)|(PUT)"},
		{"organisation_admin", "/config/individual/upload", "POST"},
		{"organisation_admin", "/config/privacy-dashboard", "GET"},
		{"organisation_admin", "/onboard/status", "GET"},
		{"user", "/onboard/password/reset", "PUT"},
		{"user", "/service/individual/record/consent-record/history", "GET"},
		{"user", "/service/idp/open-id", "GET"},
		{"user", "/service/organisation", "GET"},
		{"user", "/service/organisation/coverimage", "GET"},
		{"user", "/service/organisation/logoimage", "GET"},
		{"user", "/service/individuals", "GET"},
		{"user", "/service/individual/{individualId}", "(GET)|(PUT)"},
		{"user", "/service/image/{imageId}", "GET"},
		{"user", "/service/individual/record", "DELETE"},
		{"user", "/onboard/logout", "POST"},
		{"organisation_admin", "/onboard/logout", "POST"},
		{"audit", "/audit/consent-records", "GET"},
		{"audit", "/audit/consent-record/{consentRecordId}", "GET"},
		{"audit", "/audit/data-agreements", "GET"},
		{"audit", "/audit/data-agreement/{dataAgreementId}", "GET"},
		{"audit", "/audit/admin/logs", "GET"},
		{"config", "/config/policy", "POST"},
		{"config", "/config/policy/{policyId}", "(GET)|(PUT)|(DELETE)"},
		{"config", "/config/policy/{policyId}/revisions", "GET"},
		{"config", "/config/policies", "GET"},
		{"config", "/config/data-agreement/{dataAgreementId}", "(GET)|(PUT)|(DELETE)"},
		{"config", "/config/data-agreement", "POST"},
		{"config", "/config/data-agreements", "GET"},
		{"config", "/config/data-agreement/{dataAgreementId}/revisions", "GET"},
		{"config", "/config/data-agreement/{dataAgreementId}/revision/{revisionId}", "GET"},
		{"config", "/config/data-agreement/{dataAgreementId}/data-attributes", "GET"},
		{"config", "/config/data-agreements/data-attribute/{dataAttributeId}", "PUT"},
		{"config", "/config/data-agreements/data-attributes", "GET"},
		{"config", "/config/webhooks/event-types", "GET"},
		{"config", "/config/webhooks/payload/content-types", "GET"},
		{"config", "/config/webhooks", "GET"},
		{"config", "/config/webhook", "POST"},
		{"config", "/config/webhook/{webhookId}", "(GET)|(PUT)|(DELETE)"},
		{"config", "/config/webhook/{webhookId}/ping", "POST"},
		{"config", "/config/webhooks/{webhookId}/deliveries", "GET"},
		{"config", "/config/webhooks/{webhookId}/delivery/{deliveryId}", "GET"},
		{"config", "/config/webhooks/{webhookId}/delivery/{deliveryId}/redeliver", "POST"},
		{"config", "/config/idp/open-id", "POST"},
		{"config", "/config/idp/open-ids", "GET"},
		{"config", "/config/idp/open-id/{idpId}", "(GET)|(PUT)|(DELETE)"},
		{"config", "/config/individuals", "GET"},
		{"config", "/config/individual", "POST"},
		{"config", "/config/individual/{individualId}", "(GET)|(PUT)"},
		{"config", "/config/admin/apikey", "POST"},
		{"config", "/config/admin/apikey/{apiKeyId}", "(PUT)|(DELETE)"},
		{"config", "/config/admin/apikeys", "GET"},
		{"service", "/service/data-agreements", "GET"},
		{"service", "/service/data-agreement/{dataAgreementId}", "GET"},
		{"service", "/service/data-agreement/{dataAgreementId}/data-attributes", "GET"},
		{"service", "/service/policy/{policyId}", "GET"},
		{"service", "/service/verification/data-agreements", "GET"},
		{"service", "/service/verification/consent-record/{consentRecordId}", "GET"},
		{"service", "/service/verification/consent-records", "GET"},
		{"service", "/service/individual/record/consent-record/draft", "POST"},
		{"service", "/service/individual/record/data-agreement/{dataAgreementId}", "(GET)|(POST)"},
		{"service", "/service/individual/record/consent-record/{consentRecordId}", "PUT"},
		{"service", "/service/individual/record/consent-record", "(GET)|(POST)"},
		{"service", "/service/individual/record/consent-record/{consentRecordId}/signature", "(POST)|(PUT)"},
		{"service", "/service/individual/record/data-agreement/{dataAgreementId}/all", "GET"},
		{"service", "/service/individual/record/consent-record/history", "GET"},
		{"service", "/service/idp/open-id", "GET"},
		{"service", "/service/organisation", "GET"},
		{"service", "/service/organisation/coverimage", "GET"},
		{"service", "/service/organisation/logoimage", "GET"},
		{"service", "/service/individuals", "GET"},
		{"service", "/service/individual", "POST"},
		{"service", "/service/individual/{individualId}", "(GET)|(PUT)"},
		{"service", "/service/image/{imageId}", "GET"},
		{"service", "/service/individual/record", "DELETE"},
		{"onboard", "/onboard/organisation", "(GET)|(PUT)"},
		{"onboard", "/onboard/organisation/coverimage", "(GET)|(POST)"},
		{"onboard", "/onboard/organisation/logoimage", "(GET)|(POST)"},
		{"onboard", "/onboard/organisation", "GET"},
		{"onboard", "/onboard/password/reset", "PUT"},
		{"onboard", "/onboard/admin", "(GET)|(PUT)"},
		{"onboard", "/onboard/admin/avatarimage", "(GET)|(PUT)"},
		{"onboard", "/onboard/status", "GET"},
		{"onboard", "/onboard/logout", "POST"},
		{"config", "/config/logs/purge", "DELETE"},
		{"organisation_admin", "/config/logs/purge", "DELETE"},
	}

	for _, policy := range policies {
		if testMode {
			policy[1] = policy[1] + "/" // suffix with '/'
		} else {
			policy[1] = "/v2" + policy[1] // Prefix with '/v2'
		}
	}

	return policies
}

// CreateRbacModel
func CreateRbacModel() model.Model {
	// Initialize a new model.
	m := model.NewModel()

	// Define the request_definition section.
	m.AddDef("r", "r", "sub, obj, act")

	// Define the policy_definition section.
	m.AddDef("p", "p", "sub, obj, act")

	// Define the policy_effect section.
	m.AddDef("e", "e", "some(where (p.eft == allow))")

	// Define the matchers section.
	m.AddDef("m", "m", "r.sub == p.sub && keyMatch4(r.obj, p.obj) && (r.act == p.act || r.act == '*' || regexMatch(r.act, p.act))")

	return m
}
