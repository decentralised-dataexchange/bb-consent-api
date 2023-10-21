package http_path

// Data agreements
const ServiceReadDataAgreement = "/v2/service/data-agreement/{dataAgreementId}"
const ServiceListDataAgreements = "/v2/service/data-agreements"

// Policy
const ServiceReadPolicy = "/v2/service/policy/{policyId}"

// Data attributes
const ServiceListDataAttributesForDataAgreement = "/v2/service/data-agreement/{dataAgreementId}/data-attributes"

// Verification mechanisms
const ServiceVerificationAgreementList = "/v2/service/verification/data-agreements/"
const ServiceVerificationAgreementConsentRecordRead = "/v2/service/verification/data-agreement/"
const ServiceVerificationConsentRecordList = "/v2/service/verification/records"

// Recording consent
const ServiceCreateIndividualConsentRecord = "/v2/service/individual/data-agreement/{dataAgreementId}/record"
const ServiceUpdateIndividualConsentRecord = "/v2/service/individual/data-agreement/{dataAgreementId}"
const ServiceListIndividualRecordList = "/v2/service/individual/record/data-agreement/"
const ServiceReadIndividualRecordRead = "/v2/service/individual/record/data-agreement/{dataAgreementId}/"

// Idp
const ServiceReadIdp = "/service/idp/open-id/{idpId}"
