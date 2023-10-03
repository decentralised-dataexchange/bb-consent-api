package org

import (
	"context"
	"errors"
	"log"

	"github.com/bb-consent/api/src/database"
	"github.com/bb-consent/api/src/orgtype"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Admin Users
type Admin struct {
	UserID string
	RoleID int
}

// Organization organization data type
type Organization struct {
	ID                                primitive.ObjectID `bson:"_id,omitempty"`
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

// Lawful basis of processing IDs
const (
	ConsentBasis            = 0
	ContractBasis           = 1
	LegalObligationBasis    = 2
	VitalInterestBasis      = 3
	PublicTaskBasis         = 4
	LegitimateInterestBasis = 5
)

// LawfulBasisOfProcessingMapping Structure defining lawful basis of processing label and ID mapping
type LawfulBasisOfProcessingMapping struct {
	ID  int
	Str string
}

// LawfulBasisOfProcessingMappings List of available lawful basis of processing mappings
var LawfulBasisOfProcessingMappings = []LawfulBasisOfProcessingMapping{
	{
		ID:  ConsentBasis,
		Str: "Consent",
	},
	{
		ID:  ContractBasis,
		Str: "Contract",
	},
	{
		ID:  LegalObligationBasis,
		Str: "Legal Obligation",
	},
	{
		ID:  VitalInterestBasis,
		Str: "Vital Interest",
	},
	{
		ID:  PublicTaskBasis,
		Str: "Public Task",
	},
	{
		ID:  LegitimateInterestBasis,
		Str: "Legitimate Interest",
	},
}

func collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("organizations")
}

// Add Adds an organization
func Add(org Organization) (Organization, error) {

	org.ID = primitive.NewObjectID()
	_, err := collection().InsertOne(context.TODO(), &org)
	if err != nil {
		return org, err
	}
	return org, nil
}

// Get Gets a single organization by given id
func Get(organizationID string) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	var result Organization
	err = collection().FindOne(context.TODO(), bson.M{"_id": orgID}).Decode(&result)

	return result, err
}

// Get Gets a single organization
func GetOrganization() (Organization, error) {

	var result Organization
	err := collection().FindOne(context.TODO(), bson.M{}).Decode(&result)

	return result, err
}

// Update Updates the organization
func Update(org Organization) (Organization, error) {

	filter := bson.M{"_id": org.ID}
	update := bson.M{"$set": org}

	_, err := collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return org, err
	}
	return org, err
}

// UpdateCoverImage Update the organization image
func UpdateCoverImage(organizationID string, imageID string, imageURL string) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$set": bson.M{"coverimageid": imageID, "coverimageurl": imageURL}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// UpdateLogoImage Update the organization image
func UpdateLogoImage(organizationID string, imageID string, imageURL string) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$set": bson.M{"logoimageid": imageID, "logoimageurl": imageURL}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// AddAdminUsers Add admin users to organization
func AddAdminUsers(organizationID string, admin Admin) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$push": bson.M{"admins": admin}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// GetAdminUsers Get admin users of organization
func GetAdminUsers(organizationID string) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	filter := bson.M{"_id": orgID}
	projection := bson.M{"admins": 1}

	findOptions := options.FindOne().SetProjection(projection)

	var result Organization
	err = collection().FindOne(context.TODO(), filter, findOptions).Decode(&result)

	return result, err
}

// DeleteAdminUsers Delete admin users from organization
func DeleteAdminUsers(organizationID string, admin Admin) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$pull": bson.M{"admins": admin}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// UpdateOrganizationsOrgType Updates the embedded organization type snippet of all Organization
func UpdateOrganizationsOrgType(oType orgtype.OrgType) error {

	filter := bson.M{"type._id": oType.ID}
	update := bson.M{"$set": bson.M{"type": oType}}

	_, err := collection().UpdateMany(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	log.Println("successfully updated organiztions for type name change")
	return nil
}

// UpdatePurposes Update the organization purposes
func UpdatePurposes(organizationID string, purposes []Purpose) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$set": bson.M{"purposes": purposes}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// DeletePurposes Delete the given purpose
func DeletePurposes(organizationID string, purposes Purpose) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$pull": bson.M{"purposes": purposes}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// GetPurpose Get the organization purpose by ID
func GetPurpose(organizationID string, purposeID string) (Purpose, error) {

	o, err := Get(organizationID)
	if err != nil {
		return Purpose{}, err
	}

	for _, p := range o.Purposes {
		if p.ID == purposeID {
			return p, nil
		}
	}
	return Purpose{}, errors.New("failed to find the purpose")
}

// AddTemplates Add the organization templates
func AddTemplates(organizationID string, template Template) error {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$push": bson.M{"templates": template}})
	if err != nil {
		return err
	}
	return nil
}

// DeleteTemplates Delete the organization templates
func DeleteTemplates(organizationID string, templates Template) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$pull": bson.M{"templates": templates}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// UpdateTemplates Update the organization templates
func UpdateTemplates(organizationID string, templates []Template) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$set": bson.M{"templates": templates}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// GetTemplate Get the organization template by ID
func GetTemplate(organizationID string, templateID string) (Template, error) {

	o, err := Get(organizationID)
	if err != nil {
		return Template{}, err
	}

	for _, t := range o.Templates {
		if t.ID == templateID {
			return t, nil
		}
	}
	return Template{}, errors.New("Failed to find the template")
}

// SetEnabled Sets the enabled status to true/false
func SetEnabled(organizationID string, enabled bool) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$set": bson.M{"enabled": enabled}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// GetSubscribeMethod Get org subscribe method
func GetSubscribeMethod(orgID string) (int, error) {
	var result Organization

	orgId, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return result.Subs.Method, err
	}

	filter := bson.M{"_id": orgId}
	projection := bson.M{"subs.method": 1}

	findOptions := options.FindOne().SetProjection(projection)

	err = collection().FindOne(context.TODO(), filter, findOptions).Decode(&result)

	return result.Subs.Method, err
}

// UpdateSubscribeMethod Update subscription method
func UpdateSubscribeMethod(orgID string, method int) error {
	orgId, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": orgId}
	update := bson.M{"$set": bson.M{"subs.method": method}}

	_, err = collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// UpdateSubscribeKey Update subscription key
func UpdateSubscribeKey(orgID string, key string) error {
	orgId, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": orgId}
	update := bson.M{"$set": bson.M{"subs.key": key}}

	_, err = collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// GetSubscribeKey Update subscription token
func GetSubscribeKey(orgID string) (string, error) {
	var result Organization

	orgId, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return result.Subs.Key, err
	}

	filter := bson.M{"_id": orgId}
	projection := bson.M{"subs.key": 1}
	findOptions := options.FindOne().SetProjection(projection)

	err = collection().FindOne(context.TODO(), filter, findOptions).Decode(&result)

	return result.Subs.Key, err
}

// UpdateIdentityProviderByOrgID Update the identity provider config for org
func UpdateIdentityProviderByOrgID(organizationID string, identityProviderRepresentation IdentityProviderRepresentation) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$set": bson.M{"identityproviderrepresentation": identityProviderRepresentation}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// DeleteIdentityProviderByOrgID Delete the identity provider config for org
func DeleteIdentityProviderByOrgID(organizationID string) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$set": bson.M{"identityproviderrepresentation": nil}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// UpdateExternalIdentityProviderAvailableStatus Update the external identity provider available status for org
func UpdateExternalIdentityProviderAvailableStatus(organizationID string, availableStatus bool) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$set": bson.M{"externalidentityprovideravailable": availableStatus}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// UpdateOpenIDClientByOrgID Update OpenID config for org
func UpdateOpenIDClientByOrgID(organizationID string, openIDConfig KeycloakOpenIDClient) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$set": bson.M{"keycloakopenidclient": openIDConfig}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// DeleteOpenIDClientByOrgID Delete OpenID config for org
func DeleteOpenIDClientByOrgID(organizationID string) (Organization, error) {
	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return Organization{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgID}, bson.M{"$set": bson.M{"keycloakopenidclient": nil}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// GetName Get organization name by given id
func GetName(organizationID string) (string, error) {
	var result Organization

	orgID, err := primitive.ObjectIDFromHex(organizationID)
	if err != nil {
		return result.Name, err
	}

	filter := bson.M{"_id": orgID}
	projection := bson.M{"name": 1}
	findOptions := options.FindOne().SetProjection(projection)

	err = collection().FindOne(context.TODO(), filter, findOptions).Decode(&result)

	return result.Name, err
}
