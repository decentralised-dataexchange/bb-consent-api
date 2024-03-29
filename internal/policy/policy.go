package policy

import (
	"context"

	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/maps"
)

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("policies")
}

type Policy struct {
	Id                      string `json:"id" bson:"_id,omitempty"`
	Name                    string `json:"name" valid:"required"`
	Version                 string `json:"version"`
	Url                     string `json:"url" valid:"required"`
	Jurisdiction            string `json:"jurisdiction"`
	IndustrySector          string `json:"industrySector"`
	DataRetentionPeriodDays int    `json:"dataRetentionPeriodDays"`
	GeographicRestriction   string `json:"geographicRestriction"`
	StorageLocation         string `json:"storageLocation"`
	ThirdPartyDataSharing   bool   `json:"thirdPartyDataSharing"`
	OrganisationId          string `json:"-"`
	IsDeleted               bool   `json:"-"`
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
func (prepo *PolicyRepository) Get(policyId string) (Policy, error) {

	filter := CombineFilters(prepo.DefaultFilter, bson.M{"_id": policyId})

	var result Policy
	err := Collection().FindOne(context.TODO(), filter).Decode(&result)

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

// Get Gets count of policy for organisation
func (prepo *PolicyRepository) GetPolicyCountByOrganisation() (int64, error) {

	filter := prepo.DefaultFilter

	count, err := Collection().CountDocuments(context.TODO(), filter)

	return count, err
}

// CreatePipelineForFilteringPolicies This pipeline is used for filtering policies
func CreatePipelineForFilteringPolicies(organisationId string) ([]primitive.M, error) {

	var pipeline []bson.M

	// Stage 1 - Match by `organisationId` and `isDeleted=false`
	pipeline = append(pipeline, bson.M{"$match": bson.M{"organisationid": organisationId, "isdeleted": false}})

	// Stage 2 - Lookup revision by `policyId`
	// This is done to obtain timestamp for the latest revision of the policies.
	pipeline = append(pipeline, bson.M{"$lookup": bson.M{
		"from": "revisions",
		"let":  bson.M{"localId": "$_id"},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$and": bson.A{
							bson.M{"$eq": []interface{}{"$objectid", bson.M{"$toString": "$$localId"}}},
							bson.M{"$eq": []interface{}{"$schemaname", config.Policy}},
						},
					},
				},
			},
			bson.M{
				"$sort": bson.M{"timestamp": -1},
			},
			bson.M{"$limit": int64(1)},
		},
		"as": "revisions",
	}})

	// Stage 3 - Add the timestamp from revisions
	pipeline = append(pipeline, bson.M{"$addFields": bson.M{"timestamp": bson.M{
		"$let": bson.M{
			"vars": bson.M{
				"first": bson.M{
					"$arrayElemAt": bson.A{"$revisions", 0},
				},
			},
			"in": "$$first.timestamp",
		},
	}}})

	// Stage 4 - Remove revisions field
	pipeline = append(pipeline, bson.M{
		"$project": bson.M{
			"revisions": 0,
		},
	})

	return pipeline, nil
}

// GetFirstPolicy
func (prepo *PolicyRepository) GetFirstPolicy() (Policy, error) {

	var result Policy
	err := Collection().FindOne(context.TODO(), prepo.DefaultFilter).Decode(&result)

	return result, err
}

// DeleteAllPolicies
func DeleteAllPolicies(organisationId string) error {

	_, err := Collection().DeleteMany(context.TODO(), bson.M{"organisationid": organisationId})

	return err
}
