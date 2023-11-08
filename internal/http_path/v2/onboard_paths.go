package http_path

// login
const LoginAdminUser = "/v2/onboard/admin/login"
const LoginUser = "/v2/onboard/individual/login"
const OnboardLogoutUser = "/v2/onboard/logout"

const OnboardResetPassword = "/v2/onboard/password/reset"
const OnboardForgotPassword = "/v2/onboard/password/forgot"

const ValidateUserEmail = "/v2/onboard/validate/email"
const ValidatePhoneNumber = "/v2/onboard/validate/phone"
const VerifyPhoneNumber = "/v2/onboard/verify/phone"
const VerifyOtp = "/v2/onboard/verify/otp"

const OnboardRefreshToken = "/v2/onboard/token/refresh"
const ExchangeAuthorizationCode = "/v2/onboard/token/exchange"

const GetOrganizationByID = "/v2/onboard/organisation"
const UpdateOrganization = "/v2/onboard/organisation"
const UpdateOrganizationCoverImage = "/v2/onboard/organisation/coverimage"
const UpdateOrganizationLogoImage = "/v2/onboard/organisation/logoimage"
const GetOrganizationCoverImage = "/v2/onboard/organisation/coverimage"
const GetOrganizationLogoImage = "/v2/onboard/organisation/logoimage"

const OnboardReadOrganisationAdmin = "/v2/onboard/admin"
const OnboardUpdateOrganisationAdmin = "/v2/onboard/admin"
const OnboardReadOrganisationAdminAvatar = "/v2/onboard/admin/avatarimage"
const OnboardUpdateOrganisationAdminAvatar = "/v2/onboard/admin/avatarimage"

const OnboardReadStatus = "/v2/onboard/status"
