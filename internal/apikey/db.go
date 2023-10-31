package apikey

import (
	"context"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("apiKeys")
}

type ApiKeyRepository struct {
	DefaultFilter bson.M
}

// Init
func (apiKeyRepo *ApiKeyRepository) Init(organisationId string) {
	apiKeyRepo.DefaultFilter = bson.M{"organisationid": organisationId, "isdeleted": false}
}

// Add Adds the data agreement to the db
func (apiKeyRepo *ApiKeyRepository) Add(apiKey ApiKey) (ApiKey, error) {

	_, err := Collection().InsertOne(context.TODO(), apiKey)
	if err != nil {
		return ApiKey{}, err
	}

	return apiKey, nil
}

// Get Gets a single data agreement by given id
func (apiKeyRepo *ApiKeyRepository) Get(apiKeyID string) (ApiKey, error) {
	apiKeyId, err := primitive.ObjectIDFromHex(apiKeyID)
	if err != nil {
		return ApiKey{}, err
	}

	filter := common.CombineFilters(apiKeyRepo.DefaultFilter, bson.M{"_id": apiKeyId})

	var result ApiKey
	err = Collection().FindOne(context.TODO(), filter).Decode(&result)
	return result, err
}

// Update Updates the data agreement
func (apiKeyRepo *ApiKeyRepository) Update(apiKey ApiKey) (ApiKey, error) {

	filter := common.CombineFilters(apiKeyRepo.DefaultFilter, bson.M{"_id": apiKey.Id})
	update := bson.M{"$set": apiKey}

	_, err := Collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return ApiKey{}, err
	}
	return apiKey, nil
}
