package config

// All http response content types
const (
	ContentTypeHeader         = "Content-Type"
	ContentTypeJSON           = "application/json"
	ContentTypeImage          = "image/jpeg"
	ContentTypeFormURLEncoded = "application/x-www-form-urlencoded"
)

// Application mode
const (
	SingleTenant = "single-tenant"
	MultiTenant  = "multi-tenant"
)

// All http path url variables
const (
	OrganizationId    = "organizationId"
	DataAgreementId   = "dataAgreementId"
	DataAttributeId   = "dataAttributeId"
	WebhookId         = "webhookId"
	WebhookDeliveryId = "deliveryId"
	PolicyId          = "policyId"
)
