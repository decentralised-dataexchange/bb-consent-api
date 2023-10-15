package http_path

// Global policy configuration
const UpdateGlobalPolicyConfigurations = "/v2/config/policy"
const GetGlobalPolicyConfigurations = "/v2/config/policy/{policyId}"
const UpdateGlobalPolicyConfigurationById = "/v2/config/policy/{policyId}"
const OrgListPolicyRevisions = "/v2/config/policy/{policyId}/revisions"
const OrgDeletePolicy = "/v2/config/policy/{policyId}"
const OrgListPolicy = "/v2/config/policies"

// Data agreements
const GetDataAgreementById = "/v2/config/data-agreement/{dataAgreementId}"
const AddDataAgreement = "/v2/config/data-agreement/"
const UpdateDataAgreement = "/v2/config/data-agreement/{dataAgreementId}"
const DeleteDataAgreement = "/v2/config/data-agreement/{dataAgreementId}"
const ListDataAgreements = "/v2/config/data-agreements"
const ListDataAgreementRevisions = "/v2/config/data-agreement/{dataAgreementId}/revisions"
const ReadDataAgreementRevision = "/v2/config/data-agreement/{dataAgreementId}/revision/{revisionId}"

// Data attributes
const GetDataAttributes = "/v2/config/data-agreements/data-attributes"
const AddDataAttribute = "/v2/config/data-agreements/data-attribute"
const UpdateDataAttributeById = "/v2/config/data-agreements/data-attribute/{dataAttributeId}"
const DeleteDataAttributeById = "/v2/config/data-agreements/data-attribute/{dataAttributeId}"

// Webhooks
const GetWebhookEventTypes = "/v2/config/webhooks/event-types"
const GetWebhookPayloadContentTypes = "/v2/config/webhooks/payload/content-types"
const GetAllWebhooks = "/v2/config/webhooks"
const CreateWebhook = "/v2/config/webhook"
const UpdateWebhook = "/v2/config/webhook/{webhookId}"
const DeleteWebhook = "/v2/config/webhook/{webhookId}"
const PingWebhook = "/v2/config/webhook/{webhookId}/ping"
const GetRecentWebhookDeliveries = "/v2/config/webhooks/{webhookId}/delivery"
const GetRecentWebhookDeliveryById = "/v2/config/webhooks/{webhookId}/delivery/{deliveryId}"
const RedeliverWebhookPayloadByDeliveryID = "/v2/config/webhooks/{webhookId}/delivery/{deliveryId}/redeliver"

// Organisation identity provider related API(s)
const AddIdentityProvider = "/v2/config/idp/open-id"
const UpdateIdentityProvider = "/v2/config/idp/open-id"
const DeleteIdentityProvider = "/v2/config/idp/open-id"
const GetIdentityProvider = "/v2/config/idp/open-id"

// Individuals
const GetOrganizationUsers = "/v2/config/individuals"
const RegisterUser = "/v2/config/individual"
const GetUser = "/v2/config/individual/{individualId}"
const DeleteUser = "/v2/config/individual/{individualId}"
const UpdateUser = "/v2/config/individual/{individualId}"

// Api key
const CreateAPIKey = "/v2/config/admin/apikey"
const DeleteAPIKey = "/v2/config/admin/apikey"
const GetAPIKey = "/v2/config/admin/apikey"
