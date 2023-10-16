package policy

import (
	"context"

	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/maps"
)

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("policies")
}

type Policy struct {
	Id                      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name                    string             `json:"name" valid:"required"`
	Version                 string             `json:"version"`
	Url                     string             `json:"url" valid:"required"`
	Jurisdiction            string             `json:"jurisdiction"`
	IndustrySector          string             `json:"industrySector"`
	DataRetentionPeriodDays int                `json:"dataRetentionPeriod"`
	GeographicRestriction   string             `json:"geographicRestriction"`
	StorageLocation         string             `json:"storageLocation"`
	OrganisationId          string             `json:"-"`
	IsDeleted               bool               `json:"-"`
}

func CombineFilters(filter1 bson.M, filter2 bson.M) bson.M {
	maps.Copy(filter1, filter2)
	return filter1
}

type PolicyRepository struct {
	DefaultFilter bson.M
}

// Init
func (prepo *PolicyRepository) Init(organisationId string) {
	prepo.DefaultFilter = bson.M{"organisationid": organisationId, "isdeleted": false}
}

// Add Adds the policy to the db
func (prepo *PolicyRepository) Add(policy Policy) (Policy, error) {

	_, err := Collection().InsertOne(context.TODO(), policy)
	if err != nil {
		return Policy{}, err
	}

	return policy, nil
}

// Get Gets a single policy by given id
func (prepo *PolicyRepository) Get(policyID string) (Policy, error) {
	policyId, err := primitive.ObjectIDFromHex(policyID)
	if err != nil {
		return Policy{}, err
	}

	filter := CombineFilters(prepo.DefaultFilter, bson.M{"_id": policyId})

	var result Policy
	err = Collection().FindOne(context.TODO(), filter).Decode(&result)

	return result, err
}

// Update Updates the policy
func (prepo *PolicyRepository) Update(policy Policy) (Policy, error) {

	filter := CombineFilters(prepo.DefaultFilter, bson.M{"_id": policy.Id})
	update := bson.M{"$set": policy}

	_, err := Collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return policy, err
	}
	return policy, err
}
