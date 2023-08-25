package org

import (
	"github.com/bb-consent/api/src/database"
	"github.com/bb-consent/api/src/orgtype"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Admin Users
type Admin struct {
	UserID string
	RoleID int
}

// Organization organization data type
type Organization struct {
	ID                                bson.ObjectId `bson:"_id,omitempty"`
	Name                              string
	CoverImageID                      string
	CoverImageURL                     string
	LogoImageID                       string
	LogoImageURL                      string
	Location                          string
	Type                              orgtype.OrgType
	Jurisdiction                      string
	Disclosure                        string
	Restriction                       string
	Shared3PP                         bool
	Description                       string
	Enabled                           bool
	PolicyURL                         string
	EulaURL                           string
	Templates                         []Template
	Purposes                          []Purpose
	Admins                            []Admin
	Subs                              Subscribe
	HlcSupport                        bool
	DataRetention                     DataRetention
	IdentityProviderRepresentation    IdentityProviderRepresentation
	KeycloakOpenIDClient              KeycloakOpenIDClient
	ExternalIdentityProviderAvailable bool
}

// Template Template stored as part of the org
type Template struct {
	ID           string
	Consent      string
	PurposeIDs   []string
	DataExchange bool
	Description  string
}

// Purpose data type
type Purpose struct {
	ID                        string
	Name                      string
	Description               string
	LawfulUsage               bool
	LawfulBasisOfProcessing   int
	PolicyURL                 string
	AttributeType             int
	Jurisdiction              string
	Disclosure                string
	IndustryScope             string
	DataRetention             DataRetention
	Restriction               string
	Shared3PP                 bool
	SSIID                     string
	CloudAgentDataAgreementId string
	Version                   string
	PublishFlag               bool
}

// DataRetention data retention configuration
type DataRetention struct {
	RetentionPeriod int64
	Enabled         bool
}

// Subscribe Defines how users can subscribe to organization
type Subscribe struct {
	Method int
	Key    string
}

// IdentityProviderRepresentation Request body that describes identity provider
type IdentityProviderRepresentation struct {
	ProviderID                string                       `json:"providerId"`
	Config                    IdentityProviderOpenIDConfig `json:"config"`
	Alias                     string                       `json:"alias" valid:"required"`
	StoreToken                bool                         `json:"storeToken"`
	AddReadTokenRoleOnCreate  bool                         `json:"addReadTokenRoleOnCreate"`
	Enabled                   bool                         `json:"enabled"`
	FirstBrokerLoginFlowAlias string                       `json:"firstBrokerLoginFlowAlias"`
	LinkOnly                  bool                         `json:"linkOnly"`
	PostBrokerLoginFlowAlias  string                       `json:"postBrokerLoginFlowAlias"`
	TrustEmail                bool                         `json:"trustEmail"`
	AuthenticateByDefault     bool                         `json:"authenticateByDefault"`
}

// IdentityProviderOpenIDConfig Request body that describes identity provider OpenID config
type IdentityProviderOpenIDConfig struct {
	AuthorizationURL     string `json:"authorizationUrl" valid:"required"`
	TokenURL             string `json:"tokenUrl" valid:"required"`
	LogoutURL            string `json:"logoutUrl"`
	ClientAuthMethod     string `json:"clientAuthMethod"`
	SyncMode             string `json:"syncMode"`
	ClientID             string `json:"clientId" valid:"required"`
	ClientSecret         string `json:"clientSecret" valid:"required"`
	JWKSURL              string `json:"jwksUrl"`
	UserInfoURL          string `json:"userInfoUrl"`
	DefaultScope         string `json:"defaultScope"`
	ValidateSignature    bool   `json:"validateSignature"`
	BackchannelSupported bool   `json:"backchannelSupported"`
	DisableUserInfo      bool   `json:"disableUserInfo"`
	HideOnLoginPage      bool   `json:"hideOnLoginPage"`
	Issuer               string `json:"issuer"`
	UseJWKSURL           bool   `json:"useJwksUrl"`
}

// KeycloakOpenIDClient Describes OpenID client for managing external IDP login sessions
type KeycloakOpenIDClient struct {
	ClientID                           string                                                 `json:"clientId"`
	SurrogateAuthRequired              bool                                                   `json:"surrogateAuthRequired"`
	Enabled                            bool                                                   `json:"enabled"`
	AlwaysDisplayInConsole             bool                                                   `json:"alwaysDisplayInConsole"`
	ClientAuthenticatorType            string                                                 `json:"clientAuthenticatorType"`
	RedirectUris                       []string                                               `json:"redirectUris"`
	WebOrigins                         []string                                               `json:"webOrigins"`
	NotBefore                          int                                                    `json:"notBefore"`
	BearerOnly                         bool                                                   `json:"bearerOnly"`
	ConsentRequired                    bool                                                   `json:"consentRequired"`
	StandardFlowEnabled                bool                                                   `json:"standardFlowEnabled"`
	ImplicitFlowEnabled                bool                                                   `json:"implicitFlowEnabled"`
	DirectAccessGrantsEnabled          bool                                                   `json:"directAccessGrantsEnabled"`
	ServiceAccountsEnabled             bool                                                   `json:"serviceAccountsEnabled"`
	PublicClient                       bool                                                   `json:"publicClient"`
	FrontchannelLogout                 bool                                                   `json:"frontchannelLogout"`
	Protocol                           string                                                 `json:"protocol"`
	Attributes                         KeycloakOpenIDClientAttributes                         `json:"attributes"`
	AuthenticationFlowBindingOverrides KeycloakOpenIDClientAuthenticationFlowBindingOverrides `json:"authenticationFlowBindingOverrides"`
	FullScopeAllowed                   bool                                                   `json:"fullScopeAllowed"`
	NodeReRegistrationTimeout          int                                                    `json:"nodeReRegistrationTimeout"`
	DefaultClientScopes                []string                                               `json:"defaultClientScopes"`
	OptionalClientScopes               []string                                               `json:"optionalClientScopes"`
	Access                             KeycloakOpenIDClientAccess                             `json:"access"`
}

// KeycloakOpenIDClientAttributes Describes OpenID client attributes
type KeycloakOpenIDClientAttributes struct {
	SamlAssertionSignature                string `json:"saml.assertion.signature"`
	SamlForcePostBinding                  string `json:"saml.force.post.binding"`
	SamlMultivaluedRoles                  string `json:"saml.multivalued.roles"`
	SamlEncrypt                           string `json:"saml.encrypt"`
	BackchannelLogoutRevokeOfflineTokens  string `json:"backchannel.logout.revoke.offline.tokens"`
	SamlServerSignature                   string `json:"saml.server.signature"`
	SamlServerSignatureKeyinfoExt         string `json:"saml.server.signature.keyinfo.ext"`
	ExcludeSessionStateFromAuthResponse   string `json:"exclude.session.state.from.auth.response"`
	BackchannelLogoutSessionRequired      string `json:"backchannel.logout.session.required"`
	BackchannelLogoutURL                  string `json:"backchannel.logout.url"`
	ClientCredentialsUseRefreshToken      string `json:"client_credentials.use_refresh_token"`
	SamlForceNameIDFormat                 string `json:"saml_force_name_id_format"`
	SamlClientSignature                   string `json:"saml.client.signature"`
	TLSClientCertificateBoundAccessTokens string `json:"tls.client.certificate.bound.access.tokens"`
	SamlAuthnstatement                    string `json:"saml.authnstatement"`
	DisplayOnConsentScreen                string `json:"display.on.consent.screen"`
	SamlOnetimeuseCondition               string `json:"saml.onetimeuse.condition"`
}

// KeycloakOpenIDClientAuthenticationFlowBindingOverrides Describes OpenID client authentication flow binding overrides
type KeycloakOpenIDClientAuthenticationFlowBindingOverrides struct {
	DirectGrant string `json:"direct_grant"`
	Browser     string `json:"browser"`
}

// KeycloakOpenIDClientAccess Describes OpenID client access config
type KeycloakOpenIDClientAccess struct {
	View      bool `json:"view"`
	Configure bool `json:"configure"`
	Manage    bool `json:"manage"`
}

func session() *mgo.Session {
	return database.DB.Session.Copy()
}

func collection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("organizations")
}

// Add Adds an organization
func Add(org Organization) (Organization, error) {
	s := session()
	defer s.Close()

	org.ID = bson.NewObjectId()
	return org, collection(s).Insert(&org)
}

// Get Gets a single organization by given id
func Get(organizationID string) (Organization, error) {
	s := session()
	defer s.Close()

	var result Organization
	err := collection(s).FindId(bson.ObjectIdHex(organizationID)).One(&result)

	return result, err
}

// Update Updates the organization
func Update(org Organization) (Organization, error) {
	s := session()
	defer s.Close()

	err := collection(s).UpdateId(org.ID, org)
	return org, err
}

// UpdateCoverImage Update the organization image
func UpdateCoverImage(organizationID string, imageID string, imageURL string) (Organization, error) {
	s := session()
	defer s.Close()

	err := collection(s).Update(bson.M{"_id": bson.ObjectIdHex(organizationID)}, bson.M{"$set": bson.M{"coverimageid": imageID, "coverimageurl": imageURL}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// UpdateLogoImage Update the organization image
func UpdateLogoImage(organizationID string, imageID string, imageURL string) (Organization, error) {
	s := session()
	defer s.Close()

	err := collection(s).Update(bson.M{"_id": bson.ObjectIdHex(organizationID)}, bson.M{"$set": bson.M{"logoimageid": imageID, "logoimageurl": imageURL}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// AddAdminUsers Add admin users to organization
func AddAdminUsers(organizationID string, admin Admin) (Organization, error) {
	s := session()
	defer s.Close()

	err := collection(s).Update(bson.M{"_id": bson.ObjectIdHex(organizationID)}, bson.M{"$push": bson.M{"admins": admin}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// GetAdminUsers Get admin users of organization
func GetAdminUsers(organizationID string) (Organization, error) {
	s := session()
	defer s.Close()

	var result Organization
	err := collection(s).FindId(bson.ObjectIdHex(organizationID)).Select(bson.M{"admins": 1}).One(&result)

	return result, err
}

// DeleteAdminUsers Delete admin users from organization
func DeleteAdminUsers(organizationID string, admin Admin) (Organization, error) {
	s := session()
	defer s.Close()

	err := collection(s).Update(bson.M{"_id": bson.ObjectIdHex(organizationID)}, bson.M{"$pull": bson.M{"admins": admin}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}
