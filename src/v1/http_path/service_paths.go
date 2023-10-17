package http_path

const GetUserConsentHistory = "/v1/users/{userID}/consenthistory"
const GetConsentPurposeByID = "/v1/organizations/{orgID}/users/{userID}/consents/{consentID}/purposes/{purposeID}"
const GetConsentByID = "/v1/organizations/{orgID}/users/{userID}/consents/{consentID}"
const GetConsents = "/v1/organizations/{orgID}/users/{userID}/consents"
const GetPurposeAllConsentStatus = "/v1/organizations/{orgID}/users/{userID}/consents/{consentID}/purposes/{purposeID}/status"
const UpdatePurposeAllConsentsv2 = "/v1/organizations/{orgID}/users/{userID}/consents/{consentID}/purposes/{purposeID}/status"
const UpdatePurposeAttribute = "/v1/organizations/{orgID}/users/{userID}/consents/{consentID}/purposes/{purposeID}/attributes/{attributeID}"
const GetMyOrgDataRequestStatus = "/v1/user/organizations/{organizationID}/data-status"
const GetDeleteMyData = "/v1/user/organizations/{orgID}/data-delete"
const DeleteMyData = "/v1/user/organizations/{orgID}/data-delete"
const GetDeleteMyDataStatus = "/v1/user/organizations/{orgID}/data-delete/status"
const DataDeleteCancelMyDataRequest = "/v1/user/organizations/{orgID}/data-delete/{dataReqID}/cancel"
const GetDownloadMyData = "/v1/user/organizations/{orgID}/data-download"
const DownloadMyData = "/v1/user/organizations/{orgID}/data-download"
const GetDownloadMyDataStatus = "/v1/user/organizations/{orgID}/data-download/status"
const DataDownloadCancelMyDataRequest = "/v1/user/organizations/{orgID}/data-download/{dataReqID}/cancel"
const GetUserOrgsAndConsents = "/v1/GetUserOrgsAndConsents"