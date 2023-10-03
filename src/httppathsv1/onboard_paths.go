package httppathsv1

const AddOrganization = "/v1/organizations"
const GetOrganizationRoles = "/v1/organizations/roles"
const GetSubscribeMethods = "/v1/organizations/subscribe-methods"
const GetDataRequestStatus = "/v1/organizations/data-requests"
const GetOrganizationTypes = "/v1/organizations/types"
const AddOrganizationType = "/v1/organizations/types"
const UpdateOrganizationType = "/v1/organizations/types/{typeID}"
const DeleteOrganizationType = "/v1/organizations/types/{typeID}"
const GetOrganizationTypeByID = "/v1/organizations/types/{typeID}"
const UpdateOrganizationTypeImage = "/v1/organizations/types/{typeID}/image"
const GetOrganizationTypeImage = "/v1/organizations/types/{typeID}/image"

const GetWebhookEventTypes = "/v1/organizations/webhooks/event-types"

const GetOrganizationByID = "/v1/organizations/{organizationID}"
const UpdateOrganization = "/v1/organizations/{organizationID}"
const UpdateOrganizationCoverImage = "/v1/organizations/{organizationID}/coverimage"
const UpdateOrganizationLogoImage = "/v1/organizations/{organizationID}/logoimage"
const GetOrganizationImage = "/v1/organizations/{organizationID}/image/{imageID}"
const GetOrganizationImageWeb = "/v1/organizations/{organizationID}/image/{imageID}/web"

const UpdateOrgEula = "/v1/organizations/{organizationID}/eulaURL"
const DeleteOrgEula = "/v1/organizations/{organizationID}/eulaURL"

const AddOrgAdmin = "/v1/organizations/{organizationID}/admins"
const GetOrgAdmins = "/v1/organizations/{organizationID}/admins"
const DeleteOrgAdmin = "/v1/organizations/{organizationID}/admins"

// Organisation identity provider related API(s)
const AddIdentityProvider = "/v1/organizations/{organizationID}/idp/open-id"
const UpdateIdentityProvider = "/v1/organizations/{organizationID}/idp/open-id"
const DeleteIdentityProvider = "/v1/organizations/{organizationID}/idp/open-id"
const GetIdentityProvider = "/v1/organizations/{organizationID}/idp/open-id"

// Login
const RegisterUser = "/v1/users/register"
const LoginUser = "/v1/users/login"
const LoginUserV11 = "/v1.1/users/login"
const ValidateUserEmail = "/v1/users/validate/email"
const ValidatePhoneNumber = "/v1/users/validate/phone"
const VerifyPhoneNumber = "/v1/users/verify/phone"
const VerifyOtp = "/v1/users/verify/otp"

// Admin login
const LoginAdminUser = "/v1/users/admin/login"
const GetToken = "/v1/users/token"
const ResetPassword = "/v1/user/password/reset"
const ForgotPassword = "/v1/user/password/forgot"
const LogoutUser = "/v1/users/logout"
const UnregisterUser = "/v1/users/unregister"

const GetCurrentUser = "/v1/user"
const UpdateCurrentUser = "/v1/user"
const UserClientRegisterIOS = "/v1/user/register/ios"
const UserClientRegisterAndroid = "/v1/user/register/android"

const CreateAPIKey = "/v1/user/apikey"
const DeleteAPIKey = "/v1/user/apikey/revoke"
const GetAPIKey = "/v1/user/apikey"

const EnableOrganizationSubscription = "/v1/organizations/{organizationID}/subscription/enable"
const DisableOrganizationSubscription = "/v1/organizations/{organizationID}/subscription/disable"
const GetSubscribeMethod = "/v1/organizations/{organizationID}/subscribe-method"
const SetSubscribeMethod = "/v1/organizations/{organizationID}/subscribe-method"
const GetSubscribeKey = "/v1/organizations/{organizationID}/subscribe-key"
const RenewSubscribeKey = "/v1/organizations/{organizationID}/subscribe-key/renew"
const GetOrganizationSubscriptionStatus = "/v1/organizations/{organizationID}/subscription"

const GetDataRequests = "/v1/organizations/{orgID}/data-requests"
const GetDataRequest = "/v1/organizations/{orgID}/data-requests/{dataReqID}"
const UpdateDataRequests = "/v1/organizations/{orgID}/data-requests/{dataReqID}"

const NotifyDataBreach = "/v1/organizations/{orgID}/notify-data-breach"
const NotifyEvents = "/v1/organizations/{orgID}/notify-events"

const AddUserToOrganization = "/v1/organizations/{organizationID}/users"
const DeleteUserFromOrganization = "/v1/organizations/{organizationID}/users/{userID}"
const GetOrganizationUsers = "/v1/organizations/{organizationID}/users"
const GetOrganizationUsersCount = "/v1/organizations/{organizationID}/users/count"
