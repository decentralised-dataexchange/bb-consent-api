package individual

import (
	"context"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("individuals")
}

type Individual struct {
	Id                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ExternalId         string             `json:"externalId"`
	ExternalIdType     string             `json:"externalIdType"`
	IdentityProviderId string             `json:"identityProviderId"`
	Name               string             `json:"name" valid:"required"`
	IamId              string             `json:"iamId"`
	Email              string             `json:"email" valid:"required"`
	Phone              string             `json:"phone" valid:"required"`
	IsOnboardedFromId  bool               `json:"-"`
	OrganisationId     string             `json:"-"`
	IsDeleted          bool               `json:"-"`
}

type IndividualRepository struct {
	DefaultFilter bson.M
}

// Init
func (iRepo *IndividualRepository) Init(organisationId string) {
	iRepo.DefaultFilter = bson.M{"organisationid": organisationId, "isdeleted": false}
}

// Add Adds the individual to the db
func (iRepo *IndividualRepository) Add(individual Individual) (Individual, error) {

	_, err := Collection().InsertOne(context.TODO(), individual)
	if err != nil {
		return Individual{}, err
	}

	return individual, nil
}

// Get Gets a single individual by given id
func (iRepo *IndividualRepository) Get(individualID string) (Individual, error) {
	var result Individual
	individualId, err := primitive.ObjectIDFromHex(individualID)
	if err != nil {
		return result, err
	}

	filter := common.CombineFilters(iRepo.DefaultFilter, bson.M{"_id": individualId})

	err = Collection().FindOne(context.TODO(), filter).Decode(&result)

	return result, err
}

// Update Updates the individual
func (iRepo *IndividualRepository) Update(individual Individual) (Individual, error) {

	filter := common.CombineFilters(iRepo.DefaultFilter, bson.M{"_id": individual.Id})
	update := bson.M{"$set": individual}

	_, err := Collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return individual, err
	}
	return individual, nil
}

// Get Gets a single individual by given external id
func (iRepo *IndividualRepository) GetByExternalId(externalId string) (Individual, error) {

	filter := common.CombineFilters(iRepo.DefaultFilter, bson.M{"externalid": externalId})

	var result Individual
	err := Collection().FindOne(context.TODO(), filter).Decode(&result)

	return result, err
}
