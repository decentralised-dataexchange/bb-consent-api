package policy

import (
	"context"
	"errors"

	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	Revisions               []Revision         `json:"-"`
}

// Add Adds the policy to the db
func Add(policy Policy) (Policy, error) {

	_, err := Collection().InsertOne(context.TODO(), policy)
	if err != nil {
		return Policy{}, err
	}

	return policy, nil
}

// Get Gets a single policy by given id
func Get(policyID string, organisationId string) (Policy, error) {
	policyId, err := primitive.ObjectIDFromHex(policyID)
	if err != nil {
		return Policy{}, err
	}

	var result Policy
	err = Collection().FindOne(context.TODO(), bson.M{"_id": policyId, "organisationid": organisationId, "isdeleted": false}).Decode(&result)

	return result, err
}

// GetRevisionById Get revision by id
func GetRevisionById(revisionId string, organisationId string) (Revision, error) {
	// Find the policy
	matchStage := bson.M{
		"$match": bson.M{
			"organisationid": organisationId,
			"isdeleted":      false,
			"revisions": bson.M{
				"$elemMatch": bson.M{
					"id": revisionId,
				},
			},
		},
	}
	// Select matched revision
	projectStage := bson.M{
		"$project": bson.M{ // Ensure to use '$project'
			"revisions": bson.M{
				"$filter": bson.M{
					"input": "$revisions",
					"as":    "r",
					"cond": bson.M{
						"$eq": bson.A{"$$r.id", revisionId},
					},
				},
			},
		},
	}
	pipeline := []bson.M{
		matchStage,
		projectStage,
	}

	// Execute aggregate pipeline
	cur, err := Collection().Aggregate(context.TODO(), pipeline)
	if err != nil {
		return Revision{}, err
	}
	// Close the cursor
	defer cur.Close(context.TODO())

	// Deserialise aggregate results to policies
	var policies []Policy
	if err = cur.All(context.TODO(), &policies); err != nil {
		return Revision{}, err
	}

	if len(policies[0].Revisions) == 0 {
		return Revision{}, errors.New("matching revision was not found")
	}

	return policies[0].Revisions[0], err
}

// Update Updates the policy
func Update(policy Policy, organisationId string) (Policy, error) {

	filter := bson.M{"_id": policy.Id, "organisationid": organisationId}
	update := bson.M{"$set": policy}

	_, err := Collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return policy, err
	}
	return policy, err
}
