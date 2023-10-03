package httppathsv1

// Global policy configuration
const GetGlobalPolicyConfiguration = "/v1/organizations/{organizationID}/global-policy-configuration"
const UpdateGlobalPolicyConfiguration = "/v1/organizations/{organizationID}/global-policy-configuration"

// Data agreements
const AddConsentPurposes = "/v1/organizations/{organizationID}/purposes"
const GetPurposes = "/v1/organizations/{organizationID}/purposes"
const DeleteConsentPurposeByID = "/v1/organizations/{organizationID}/purposes/{purposeID}"
const UpdatePurposeByID = "/v1/organizations/{organizationID}/purposes/{purposeID}"
const GetPurposeByID = "/v1/organizations/{organizationID}/purposes/{purposeID}"

// Data attributes
const AddConsentTemplates = "/v1/organizations/{organizationID}/templates"
const GetTemplates = "/v1/organizations/{organizationID}/templates"
const DeleteConsentTemplateByID = "/v1/organizations/{organizationID}/templates/{templateID}"
const GetTemplateByID = "/v1/organizations/{organizationID}/templates/{templateID}"
const UpdateTemplateByID = "/v1/organizations/{organizationID}/templates/{templateID}"
const DeleteConsentTemplatesByID = "/v1/organizations/{organizationID}/purposes/{purposeID}/templates"

// Webhooks
const GetWebhookPayloadContentTypes = "/v1/organizations/webhooks/payload/content-types"
const CreateWebhook = "/v1/organizations/{orgID}/webhooks"
const GetAllWebhooks = "/v1/organizations/{orgID}/webhooks"
const GetWebhook = "/v1/organizations/{orgID}/webhooks/{webhookID}"
const DeleteWebhook = "/v1/organizations/{orgID}/webhooks/{webhookID}"
const UpdateWebhook = "/v1/organizations/{orgID}/webhooks/{webhookID}"
const PingWebhook = "/v1/organizations/{orgID}/webhooks/{webhookID}/ping"
const GetRecentWebhookDeliveries = "/v1/organizations/{orgID}/webhooks/{webhookID}/delivery"
const GetWebhookDeliveryByID = "/v1/organizations/{orgID}/webhooks/{webhookID}/delivery/{deliveryID}"
const ReDeliverWebhook = "/v1/organizations/{orgID}/webhooks/{webhookID}/delivery/{deliveryID}/redeliver"

// Filtering individuals by consents
const GetAllUsersConsentedToAttribute = "/v1/organizations/{orgID}/purposes/{purposeID}/attributes/{attributeID}/consented/users"
const GetAllUsersConsentedToPurpose = "/v1/organizations/{orgID}/purposes/{purposeID}/consented/users"
