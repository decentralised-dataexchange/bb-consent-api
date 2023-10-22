package signature

import (
	"context"

	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("signatures")
}

// Add Adds the signature to the db
func Add(signature Signature) (Signature, error) {

	_, err := Collection().InsertOne(context.TODO(), signature)
	if err != nil {
		return Signature{}, err
	}

	return signature, nil
}

// Get Gets a signature by given id
func Get(signatureID string) (Signature, error) {
	signatureId, err := primitive.ObjectIDFromHex(signatureID)
	if err != nil {
		return Signature{}, err
	}

	filter := bson.M{"_id": signatureId}

	var result Signature
	err = Collection().FindOne(context.TODO(), filter).Decode(&result)

	return result, err
}

// Update Updates the signature
func Update(signature Signature) (Signature, error) {

	filter := bson.M{"_id": signature.Id}
	update := bson.M{"$set": signature}

	_, err := Collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return signature, err
	}
	return signature, nil
}
