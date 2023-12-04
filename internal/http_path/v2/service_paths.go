package http_path

// Data agreements
const ServiceReadDataAgreement = "/service/data-agreement/{dataAgreementId}"
const ServiceListDataAgreements = "/service/data-agreements"

// Policy
const ServiceReadPolicy = "/service/policy/{policyId}"

// Data attributes
const ServiceListDataAttributesForDataAgreement = "/service/data-agreement/{dataAgreementId}/data-attributes"

// Verification mechanisms
const ServiceVerificationListDataAgreements = "/service/verification/data-agreements"
const ServiceVerificationFetchDataAgreementRecord = "/service/verification/consent-record/{consentRecordId}"
const ServiceVerificationFetchDataAgreementRecords = "/service/verification/consent-records"

// Recording consent
const ServiceCreateDraftConsentRecord = "/service/individual/record/consent-record/draft"
const ServiceCreateDataAgreementRecord = "/service/individual/record/data-agreement/{dataAgreementId}"
const ServiceReadDataAgreementRecord = "/service/individual/record/data-agreement/{dataAgreementId}"
const ServiceUpdateDataAgreementRecord = "/service/individual/record/consent-record/{consentRecordId}"
const ServiceDeleteIndividualDataAgreementRecords = "/service/individual/record"
const ServiceCreatePairedDataAgreementRecord = "/service/individual/record/consent-record"

const ServiceCreateBlankSignature = "/service/individual/record/consent-record/{consentRecordId}/signature"
const ServiceUpdateSignatureObject = "/service/individual/record/consent-record/{consentRecordId}/signature"

const ServiceFetchIndividualDataAgreementRecords = "/service/individual/record/consent-record"
const ServiceFetchRecordsForDataAgreement = "/service/individual/record/data-agreement/{dataAgreementId}/all"

const ServiceFetchRecordsHistory = "/service/individual/record/consent-record/history"

// Idp
const ServiceReadIdp = "/service/idp/open-id"

// Organisation
const ServiceReadOrganisation = "/service/organisation"
const ServiceReadOrganisationLogoImage = "/service/organisation/logoimage"
const ServiceReadOrganisationCoverImage = "/service/organisation/coverimage"
const ServiceReadOrganisationImage = "/service/image/{imageId}"

// Individuals
const ServiceCreateIndividual = "/service/individual"
const ServiceReadIndividual = "/service/individual/{individualId}"
const ServiceUpdateIndividual = "/service/individual/{individualId}"
const ServiceListIndividuals = "/service/individuals"

// Data sharing
const ServiceShowDataSharingUi = "/service/data-sharing"
