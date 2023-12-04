package http_path

// Global policy configuration
const ConfigCreatePolicy = "/config/policy"
const ConfigReadPolicy = "/config/policy/{policyId}"
const ConfigUpdatePolicy = "/config/policy/{policyId}"
const ConfigDeletePolicy = "/config/policy/{policyId}"
const ConfigListPolicies = "/config/policies"
const ConfigListPolicyRevisions = "/config/policy/{policyId}/revisions"

// Data agreements
const ConfigCreateDataAgreement = "/config/data-agreement"
const ConfigReadDataAgreement = "/config/data-agreement/{dataAgreementId}"
const ConfigUpdateDataAgreement = "/config/data-agreement/{dataAgreementId}"
const ConfigDeleteDataAgreement = "/config/data-agreement/{dataAgreementId}"
const ConfigListDataAgreements = "/config/data-agreements"
const ConfigListDataAgreementRevisions = "/config/data-agreement/{dataAgreementId}/revisions"
const ConfigListDataAttributesForDataAgreement = "/config/data-agreement/{dataAgreementId}/data-attributes"

const ReadDataAgreementRevision = "/config/data-agreement/{dataAgreementId}/revision/{revisionId}"

// Data attributes
const ConfigReadDataAttribute = "/config/data-agreements/data-attribute/{dataAttributeId}"
const ConfigCreateDataAttribute = "/config/data-agreements/data-attribute"
const ConfigUpdateDataAttribute = "/config/data-agreements/data-attribute/{dataAttributeId}"
const ConfigListDataAttributeRevisions = "/config/data-agreements/data-attribute/{dataAttributeId}/revisions"
const ConfigDeleteDataAttribute = "/config/data-agreements/data-attribute/{dataAttributeId}"
const ConfigListDataAttributes = "/config/data-agreements/data-attributes"

// Webhooks
const ConfigReadWebhook = "/config/webhook/{webhookId}"
const ConfigCreateWebhook = "/config/webhook"
const ConfigUpdateWebhook = "/config/webhook/{webhookId}"
const ConfigDeleteWebhook = "/config/webhook/{webhookId}"
const ConfigListWebhooks = "/config/webhooks"
const ConfigPingWebhook = "/config/webhook/{webhookId}/ping"
const ConfigListRecentWebhookDeliveries = "/config/webhooks/{webhookId}/deliveries"
const ConfigReadRecentWebhookDelivery = "/config/webhooks/{webhookId}/delivery/{deliveryId}"
const ConfigRedeliverWebhookPayloadByDeliveryID = "/config/webhooks/{webhookId}/delivery/{deliveryId}/redeliver"
const ConfigListWebhookEventTypes = "/config/webhooks/event-types"
const ConfigListWebhookPayloadContentTypes = "/config/webhooks/payload/content-types"

// Organisation identity provider related API(s)
const AddIdentityProvider = "/config/idp/open-id"
const UpdateIdentityProvider = "/config/idp/open-id/{idpId}"
const DeleteIdentityProvider = "/config/idp/open-id/{idpId}"
const GetIdentityProvider = "/config/idp/open-id/{idpId}"
const ConfigListIdentityProviders = "/config/idp/open-ids"

// Individuals
const ConfigCreateIndividual = "/config/individual"
const ConfigReadIndividual = "/config/individual/{individualId}"
const ConfigUpdateIndividual = "/config/individual/{individualId}"
const ConfigDeleteIndividual = "/config/individual/{individualId}"
const ConfigListIndividuals = "/config/individuals"
const ConfigCreateIndividualsInBulk = "/config/individual/upload"

// Api key
const ConfigCreateApiKey = "/config/admin/apikey"
const ConfigUpdateApiKey = "/config/admin/apikey/{apiKeyId}"
const ConfigDeleteApiKey = "/config/admin/apikey/{apiKeyId}"
const ConfigListApiKey = "/config/admin/apikeys"

const ConfigReadPrivacyDashboard = "/config/privacy-dashboard"

const ConfigPurgeOrgLogs = "/config/logs/purge"
