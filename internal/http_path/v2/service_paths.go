package http_path

// Data agreements
const ServiceReadDataAgreement = "/v2/service/data-agreement/{dataAgreementId}"
const ServiceListDataAgreements = "/v2/service/data-agreements"

// Policy
const ServiceReadPolicy = "/v2/service/policy/{policyId}"

// Data attributes
const ServiceListDataAttributesForDataAgreement = "/v2/service/data-agreement/{dataAgreementId}/data-attributes"

// Verification mechanisms
const ServiceVerificationFetchAllDataAgreementRecords = "/v2/service/verification/data-agreements"
const ServiceVerificationFetchDataAgreementRecord = "/v2/service/verification/data-agreement/{dataAgreementId}"
const ServiceVerificationFetchDataAgreementRecords = "/v2/service/verification/consent-records"

// Recording consent
const ServiceCreateDraftConsentRecord = "/v2/service/individual/record/consent-record/draft"
const ServiceCreateDataAgreementRecord = "/v2/service/individual/record/data-agreement/{dataAgreementId}"
const ServiceReadDataAgreementRecord = "/v2/service/individual/record/data-agreement/{dataAgreementId}"
const ServiceUpdateDataAgreementRecord = "/v2/service/individual/record/consent-record/{consentRecordId}"
const ServiceDeleteIndividualDataAgreementRecords = "/v2/service/individual/record"
const ServiceCreatePairedDataAgreementRecord = "/v2/service/individual/record/consent-record"

const ServiceCreateBlankSignature = "/v2/service/individual/record/consent-record/{consentRecordId}/signature"
const ServiceUpdateSignatureObject = "/v2/service/individual/record/consent-record/{consentRecordId}/signature"

const ServiceFetchIndividualDataAgreementRecords = "/v2/service/individual/record/consent-record"
const ServiceFetchRecordsForDataAgreement = "/v2/service/individual/record/data-agreement/{dataAgreementId}/all"

const ServiceFetchRecordsHistory = "/v2/service/individual/record/consent-record/history"

// Idp
const ServiceReadIdp = "/v2/service/idp/open-id/{idpId}"

// Organisation
const ServiceReadOrganisation = "/v2/service/organisation"
const ServiceReadOrganisationLogoImage = "/v2/service/organisation/logoimage"
const ServiceReadOrganisationCoverImage = "/v2/service/organisation/coverimage"
const ServiceReadOrganisationImage = "/v2/service/image/{imageId}"

// Individuals
const ServiceCreateIndividual = "/v2/service/individual"
const ServiceReadIndividual = "/v2/service/individual/{individualId}"
const ServiceUpdateIndividual = "/v2/service/individual/{individualId}"
const ServiceListIndividuals = "/v2/service/individuals"

// Data sharing
const ServiceShowDataSharingUi = "/data-sharing-ui"
