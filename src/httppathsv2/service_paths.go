package httppathsv2

// Data agreements
const ServiceDataAgreementRead = "/v2/service/data-agreement/{dataAgreementId}/"

// Global policy configuration
const ServicePolicyRead = "/service/policy/{policyId}/"

// Data attributes
const ServiceGetDataAttributes = "/v2/service/data-agreements/data-attributes"

// Verification mechanisms
const ServiceVerificationAgreementList = "/v2/service/verification/data-agreements/"
const ServiceVerificationAgreementConsentRecordRead = "/v2/service/verification/data-agreement/"
const ServiceVerificationConsentRecordList = "/v2/service/verification/records"

// Recording consent
const ServiceCreateIndividualConsentRecord = "/v2/service/individual/data-agreement/{dataAgreementId}/record"
const ServiceUpdateIndividualConsentRecord = "/v2/service/individual/data-agreement/{dataAgreementId}"
const ServiceListIndividualRecordList = "/v2/service/individual/record/data-agreement/"
const ServiceReadIndividualRecordRead = "/v2/service/individual/record/data-agreement/{dataAgreementId}/"
