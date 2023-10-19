package idp

import (
	"context"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("identityProviders")
}

type IdentityProvider struct {
	Id               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	IssuerUrl        string             `json:"issuerUrl"`
	AuthorizationURL string             `json:"authorizationUrl" valid:"required"`
	TokenURL         string             `json:"tokenUrl" valid:"required"`
	LogoutURL        string             `json:"logoutUrl" valid:"required"`
	ClientID         string             `json:"clientId" valid:"required"`
	ClientSecret     string             `json:"clientSecret" valid:"required"`
	JWKSURL          string             `json:"jwksUrl" valid:"required"`
	UserInfoURL      string             `json:"userInfoUrl" valid:"required"`
	DefaultScope     string             `json:"defaultScope" valid:"required"`
	OrganisationId   string             `json:"-"`
	IsDeleted        bool               `json:"-"`
}

type IdentityProviderRepository struct {
	DefaultFilter bson.M
}

// Init
func (idpRepo *IdentityProviderRepository) Init(organisationId string) {
	idpRepo.DefaultFilter = bson.M{"organisationid": organisationId, "isdeleted": false}
}

// IsIdentityProviderExist Check if identity provider exists
func (idpRepo *IdentityProviderRepository) IsIdentityProviderExist() (int64, error) {

	filter := idpRepo.DefaultFilter

	exists, err := Collection().CountDocuments(context.TODO(), filter)
	if err != nil {
		return exists, err
	}
	return exists, nil
}

// Add Adds the identity provider to the db
func (idpRepo *IdentityProviderRepository) Add(idp IdentityProvider) (IdentityProvider, error) {

	_, err := Collection().InsertOne(context.TODO(), idp)
	if err != nil {
		return IdentityProvider{}, err
	}

	return idp, nil
}

// Update Updates the identity provider
func (idpRepo *IdentityProviderRepository) Update(idp IdentityProvider) (IdentityProvider, error) {

	filter := common.CombineFilters(idpRepo.DefaultFilter, bson.M{"_id": idp.Id})
	update := bson.M{"$set": idp}

	_, err := Collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return idp, err
	}
	return idp, nil
}

// Get Gets a single identity provider by given id
func (idpRepo *IdentityProviderRepository) Get(idpID string) (IdentityProvider, error) {
	var result IdentityProvider
	idpId, err := primitive.ObjectIDFromHex(idpID)
	if err != nil {
		return result, err
	}

	filter := common.CombineFilters(idpRepo.DefaultFilter, bson.M{"_id": idpId})

	err = Collection().FindOne(context.TODO(), filter).Decode(&result)

	return result, err
}

// Get Gets a single identity provider by given organisation id
func (idpRepo *IdentityProviderRepository) GetByOrgId() (IdentityProvider, error) {

	filter := idpRepo.DefaultFilter

	var result IdentityProvider
	err := Collection().FindOne(context.TODO(), filter).Decode(&result)

	return result, err
}
