package revision

import (
	"context"

	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("revisions")
}

// Add Adds the revision to the db
func Add(revision Revision) (Revision, error) {

	_, err := Collection().InsertOne(context.TODO(), revision)
	if err != nil {
		return Revision{}, err
	}

	return revision, nil
}

// Update Updates the revision
func Update(revision Revision) (Revision, error) {

	filter := bson.M{"_id": revision.Id}
	update := bson.M{"$set": revision}

	_, err := Collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return revision, err
	}
	return revision, err
}

// Get Gets revision by policy id
func GetLatestByPolicyId(policyId string) (Revision, error) {

	var result Revision
	opts := options.FindOne().SetSort(bson.M{"timestamp": -1})
	err := Collection().FindOne(context.TODO(), bson.M{"objectid": policyId}, opts).Decode(&result)
	if err != nil {
		return Revision{}, err
	}

	return result, err
}

// Get Gets revisions by policy id
func ListAllByPolicyId(policyId string) ([]Revision, error) {

	var results []Revision
	cursor, err := Collection().Find(context.TODO(), bson.M{"objectid": policyId})
	if err != nil {
		return []Revision{}, err
	}

	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		return []Revision{}, err
	}

	return results, err
}

// GetByRevisionId Get revision by id
func GetByRevisionId(revisionId string) (Revision, error) {

	var result Revision
	err := Collection().FindOne(context.TODO(), bson.M{"_id": revisionId}).Decode(&result)
	if err != nil {
		return Revision{}, err
	}

	return result, err
}

// Get Gets revision by data agreement id
func GetLatestByDataAgreementId(dataAgreementId string) (Revision, error) {

	var result Revision
	opts := options.FindOne().SetSort(bson.M{"timestamp": -1})
	err := Collection().FindOne(context.TODO(), bson.M{"objectid": dataAgreementId}, opts).Decode(&result)
	if err != nil {
		return Revision{}, err
	}

	return result, err
}

// Get Gets revisions by data agreement id
func ListAllByDataAgreementId(dataAgreementId string) ([]Revision, error) {

	var results []Revision
	cursor, err := Collection().Find(context.TODO(), bson.M{"objectid": dataAgreementId})
	if err != nil {
		return []Revision{}, err
	}

	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		return []Revision{}, err
	}

	return results, err
}
