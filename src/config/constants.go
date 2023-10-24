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
	OrganizationId        = "organizationId"
	DataAgreementId       = "dataAgreementId"
	DataAttributeId       = "dataAttributeId"
	WebhookId             = "webhookId"
	WebhookDeliveryId     = "deliveryId"
	PolicyId              = "policyId"
	DataAgreementRecordId = "dataAgreementRecordId"
	IndividualId          = "individualId"
	DeliveryId            = "deliveryId"
	IdpId                 = "idpId"
	ApiKeyId              = "apiKeyId"
	IndividualHeaderKey   = "X-ConsentBB-IndividualId"
	RevisionId            = "revisionId"
	LawfulBasis           = "lawfulBasis"
)

// Schemas
const (
	DataAgreement       = "dataAgreement"
	Policy              = "policy"
	DataAgreementRecord = "dataAgreementRecord"
	DataAttribute       = "dataAttribute"
)

// Data Agreement Method of Use
const (
	Null             = "null"
	DataSource       = "data_source"
	DataUsingService = "data_using_service"
)

// Lifecycle
const (
	Draft    = "draft"
	Complete = "complete"
)

// Scopes for api key
const (
	Config  = "config"
	Service = "service"
	Audit   = "audit"
	Onboard = "onboard"
)

// Data agreement record state
const (
	Unsigned = "unsigned"
	Signed   = "signed"
)
