package revision

import (
	"context"

	"github.com/bb-consent/api/internal/database"
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

// GetByRevisionId Get revision by id
func GetByRevisionId(revisionId string) (Revision, error) {
	var result Revision

	err := Collection().FindOne(context.TODO(), bson.M{"_id": revisionId}).Decode(&result)
	if err != nil {
		return Revision{}, err
	}

	return result, err
}

// GetByRevisionIdAndSchema gets revision by id and schema
func GetByRevisionIdAndSchema(revisionId string, schemaName string) (Revision, error) {
	var result Revision

	err := Collection().FindOne(context.TODO(), bson.M{"_id": revisionId, "schemaname": schemaName}).Decode(&result)
	if err != nil {
		return result, err
	}

	return result, err
}

// GetLatestByObjectIdAndSchemaName Gets latest revision by object id and schema name
func GetLatestByObjectIdAndSchemaName(objectId string, schemaName string) (Revision, error) {

	var result Revision
	opts := options.FindOne().SetSort(bson.M{"timestamp": -1})
	err := Collection().FindOne(context.TODO(), bson.M{"objectid": objectId, "schemaname": schemaName}, opts).Decode(&result)
	if err != nil {
		return Revision{}, err
	}

	return result, err
}

// ListAllByObjectIdAndSchemaName list revisions by object id and schema name
func ListAllByObjectIdAndSchemaName(objectId string, schemaName string) ([]Revision, error) {

	var results []Revision
	opts := options.Find().SetSort(bson.M{"timestamp": -1})
	cursor, err := Collection().Find(context.TODO(), bson.M{"objectid": objectId, "schemaname": schemaName}, opts)
	if err != nil {
		return []Revision{}, err
	}

	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		return []Revision{}, err
	}

	return results, err
}
