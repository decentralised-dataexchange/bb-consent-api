package http_path

// Global policy configuration
const ConfigCreatePolicy = "/v2/config/policy"
const ConfigReadPolicy = "/v2/config/policy/{policyId}"
const ConfigUpdatePolicy = "/v2/config/policy/{policyId}"
const ConfigDeletePolicy = "/v2/config/policy/{policyId}"
const ConfigListPolicies = "/v2/config/policies"
const ConfigListPolicyRevisions = "/v2/config/policy/{policyId}/revisions"

// Data agreements
const ConfigCreateDataAgreement = "/v2/config/data-agreement"
const ConfigReadDataAgreement = "/v2/config/data-agreement/{dataAgreementId}"
const ConfigUpdateDataAgreement = "/v2/config/data-agreement/{dataAgreementId}"
const ConfigDeleteDataAgreement = "/v2/config/data-agreement/{dataAgreementId}"
const ConfigListDataAgreements = "/v2/config/data-agreements"
const ConfigListDataAgreementRevisions = "/v2/config/data-agreement/{dataAgreementId}/revisions"
const ConfigListDataAttributesForDataAgreement = "/v2/config/data-agreement/{dataAgreementId}/data-attributes"

const ReadDataAgreementRevision = "/v2/config/data-agreement/{dataAgreementId}/revision/{revisionId}"

// Data attributes
const ConfigReadDataAttribute = "/v2/config/data-agreements/data-attribute/{dataAttributeId}"
const ConfigCreateDataAttribute = "/v2/config/data-agreements/data-attribute"
const ConfigUpdateDataAttribute = "/v2/config/data-agreements/data-attribute/{dataAttributeId}"
const ConfigListDataAttributeRevisions = "/v2/config/data-agreements/data-attribute/{dataAttributeId}/revisions"
const ConfigDeleteDataAttribute = "/v2/config/data-agreements/data-attribute/{dataAttributeId}"
const ConfigListDataAttributes = "/v2/config/data-agreements/data-attributes"

// Webhooks
const ConfigReadWebhook = "/v2/config/webhook/{webhookId}"
const ConfigCreateWebhook = "/v2/config/webhook"
const ConfigUpdateWebhook = "/v2/config/webhook/{webhookId}"
const ConfigDeleteWebhook = "/v2/config/webhook/{webhookId}"
const ConfigListWebhooks = "/v2/config/webhooks"
const ConfigPingWebhook = "/v2/config/webhook/{webhookId}/ping"
const ConfigListRecentWebhookDeliveries = "/v2/config/webhooks/{webhookId}/delivery"
const ConfigReadRecentWebhookDelivery = "/v2/config/webhooks/{webhookId}/delivery/{deliveryId}"
const ConfigRedeliverWebhookPayloadByDeliveryID = "/v2/config/webhooks/{webhookId}/delivery/{deliveryId}/redeliver"
const ConfigListWebhookEventTypes = "/v2/config/webhooks/event-types"
const ConfigListWebhookPayloadContentTypes = "/v2/config/webhooks/payload/content-types"

// Organisation identity provider related API(s)
const AddIdentityProvider = "/v2/config/idp/open-id"
const UpdateIdentityProvider = "/v2/config/idp/open-id"
const DeleteIdentityProvider = "/v2/config/idp/open-id"
const GetIdentityProvider = "/v2/config/idp/open-id"

// Individuals
const ConfigCreateIndividual = "/v2/config/individual"
const ConfigReadIndividual = "/v2/config/individual/{individualId}"
const ConfigUpdateIndividual = "/v2/config/individual/{individualId}"
const ConfigDeleteIndividual = "/v2/config/individual/{individualId}"
const ConfigListIndividuals = "/v2/config/individuals"

// Api key
const CreateAPIKey = "/v2/config/admin/apikey"
const DeleteAPIKey = "/v2/config/admin/apikey"
const GetAPIKey = "/v2/config/admin/apikey"
