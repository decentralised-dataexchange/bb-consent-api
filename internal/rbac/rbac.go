package rbac

import "github.com/casbin/casbin/v2/model"

// RBAC User Roles
const (
	ROLE_USER  string = "user"
	ROLE_ADMIN string = "organisation_admin"
)

// GetRbacPolicies
func GetRbacPolicies() [][]string {

	policies := [][]string{
		{"organisation_admin", "/v2/config/policy", "POST"},
		{"organisation_admin", "/v2/config/policy/{policyId}", "(GET)|(PUT)|(DELETE)"},
		{"organisation_admin", "/v2/config/policy/{policyId}/revisions", "GET"},
		{"organisation_admin", "/v2/config/policies", "GET"},
		{"organisation_admin", "/v2/config/data-agreement/{dataAgreementId}", "(GET)|(PUT)|(DELETE)"},
		{"organisation_admin", "/v2/config/data-agreement", "POST"},
		{"organisation_admin", "/v2/config/data-agreements", "GET"},
		{"organisation_admin", "/v2/config/data-agreement/{dataAgreementId}/revisions", "GET"},
		{"organisation_admin", "/v2/config/data-agreement/{dataAgreementId}/revision/{revisionId}", "GET"},
		{"organisation_admin", "/v2/config/data-agreement/{dataAgreementId}/data-attributes", "GET"},
		{"organisation_admin", "/v2/config/data-agreements/data-attribute", "POST"},
		{"organisation_admin", "/v2/config/data-agreements/data-attribute/{dataAttributeId}", "(GET)|(PUT)|(DELETE)"},
		{"organisation_admin", "/v2/config/data-agreements/data-attribute/{dataAttributeId}/revisions", "GET"},
		{"organisation_admin", "/v2/config/data-agreements/data-attributes", "GET"},
		{"organisation_admin", "/v2/config/webhooks/event-types", "GET"},
		{"organisation_admin", "/v2/config/webhooks/payload/content-types", "GET"},
		{"organisation_admin", "/v2/config/webhooks", "GET"},
		{"organisation_admin", "/v2/config/webhook", "POST"},
		{"organisation_admin", "/v2/config/webhook/{webhookId}", "(GET)|(PUT)|(DELETE)"},
		{"organisation_admin", "/v2/config/webhook/{webhookId}/ping", "POST"},
		{"organisation_admin", "/v2/config/webhooks/{webhookId}/deliveries", "GET"},
		{"organisation_admin", "/v2/config/webhooks/{webhookId}/delivery/{deliveryId}", "GET"},
		{"organisation_admin", "/v2/config/webhooks/{webhookId}/delivery/{deliveryId}/redeliver", "POST"},
		{"organisation_admin", "/v2/config/idp/open-id", "POST"},
		{"organisation_admin", "/v2/config/idp/open-ids", "GET"},
		{"organisation_admin", "/v2/config/idp/open-id/{idpId}", "(GET)|(PUT)|(DELETE)"},
		{"organisation_admin", "/v2/config/individuals", "GET"},
		{"organisation_admin", "/v2/config/individual", "POST"},
		{"organisation_admin", "/v2/config/individual/{individualId}", "(GET)|(PUT)"},
		{"organisation_admin", "/v2/config/admin/apikey", "POST"},
		{"organisation_admin", "/v2/config/admin/apikey/{apiKeyId}", "(PUT)|(DELETE)"},
		{"organisation_admin", "/v2/config/admin/apikeys", "GET"},
		{"user", "/v2/service/data-agreements", "GET"},
		{"user", "/v2/service/data-agreement/{dataAgreementId}", "GET"},
		{"user", "/v2/service/data-agreement/{dataAgreementId}/data-attributes", "GET"},
		{"user", "/v2/service/policy/{policyId}", "GET"},
		{"user", "/v2/service/verification/data-agreements", "GET"},
		{"user", "/v2/service/verification/data-agreement/{dataAgreementId}", "GET"},
		{"user", "/v2/service/verification/consent-records", "GET"},
		{"user", "/v2/service/individual/record/consent-record/draft", "POST"},
		{"user", "/v2/service/individual/record/data-agreement/{dataAgreementId}", "(GET)|(POST)"},
		{"user", "/v2/service/individual/record/consent-record/{consentRecordId}", "PUT"},
		{"user", "/v2/service/individual/record/consent-record", "(GET)|(POST)"},
		{"user", "/v2/service/individual/record/consent-record/{consentRecordId}/signature", "(POST)|(PUT)"},
		{"user", "/v2/service/individual/record/data-agreement/{dataAgreementId}/all", "GET"},
		{"organisation_admin", "/v2/audit/consent-records", "GET"},
		{"organisation_admin", "/v2/audit/consent-record/{consentRecordId}", "GET"},
		{"organisation_admin", "/v2/audit/data-agreements", "GET"},
		{"organisation_admin", "/v2/audit/data-agreement/{dataAgreementId}", "GET"},
		{"organisation_admin", "/v2/audit/admin/logs", "GET"},
		{"organisation_admin", "/v2/onboard/organisation", "(GET)|(PUT)"},
		{"organisation_admin", "/v2/onboard/organisation/coverimage", "(GET)|(POST)"},
		{"organisation_admin", "/v2/onboard/organisation/logoimage", "(GET)|(POST)"},
		{"user", "/v2/onboard/organisation", "GET"},
		{"user", "/v2/onboard/organisation/coverimage", "GET"},
		{"user", "/v2/onboard/organisation/logoimage", "GET"},
		{"organisation_admin", "/v2/onboard/password/reset", "PUT"},
		{"organisation_admin", "/v2/onboard/admin", "(GET)|(PUT)"},
		{"organisation_admin", "/v2/onboard/admin/avatarimage", "(GET)|(PUT)"},
		{"organisation_admin", "/v2/config/individual/upload", "POST"},
		{"organisation_admin", "/v2/config/privacy-dashboard", "GET"},
		{"organisation_admin", "/v2/onboard/status", "GET"},
		{"user", "/v2/onboard/password/reset", "PUT"},
		{"user", "/v2/service/individual/record/consent-record/history", "GET"},
		{"user", "/v2/service/idp/open-id/{idpId}", "GET"},
		{"user", "/v2/service/organisation", "GET"},
		{"user", "/v2/service/organisation/coverimage", "GET"},
		{"user", "/v2/service/organisation/logoimage", "GET"},
		{"user", "/v2/service/individuals", "GET"},
		{"user", "/v2/service/individual", "POST"},
		{"user", "/v2/service/individual/{individualId}", "(GET)|(PUT)"},
		{"user", "/v2/service/image/{imageId}", "GET"},
		{"user", "/v2/service/individual/record", "DELETE"},
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
