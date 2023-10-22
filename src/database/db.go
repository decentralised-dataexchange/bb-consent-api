package database

import (
	"context"
	"log"
	"time"

	"github.com/bb-consent/api/src/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type db struct {
	Client *mongo.Client
	Name   string
}

// DB Database session pointer
var DB db

// Init Connects to the DB, initializes the collection
func Init(config *config.Configuration) error {
	MongoDBURL := "mongodb://" + config.DataBase.UserName + ":" + config.DataBase.Password + "@" + config.DataBase.Hosts[0] + "/" + config.DataBase.Name

	clientOptions := options.Client().ApplyURI(MongoDBURL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a new MongoDB client
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v", err)
		return err
	}

	// Ping the MongoDB server
	err = client.Ping(ctx, nil)
	if err != nil {
		return err
	}

	DB = db{
		Client: client,
		Name:   config.DataBase.Name,
	}

	err = initCollection("organizations", []string{"name"}, true)
	if err != nil {
		log.Printf("initialising collection: %v", err)
		return err
	}

	err = initCollection("users", []string{"_id", "email", "phone"}, true)
	if err != nil {
		return err
	}

	err = initCollection("orgTypes", []string{"type"}, true)
	if err != nil {
		return err
	}

	err = initCollection("consents", []string{"userid", "orgid"}, false)
	if err != nil {
		return err
	}

	//TODO: Set unique to false
	err = initCollection("images", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("otps", []string{"phone"}, true)
	if err != nil {
		return err
	}
	err = initCollection("notifications", []string{"userid"}, false)
	if err != nil {
		return err
	}

	err = initCollection("consentHistory", []string{"userid"}, false)
	if err != nil {
		return err
	}

	err = initCollection("misc", []string{"type"}, false)
	if err != nil {
		return err
	}

	err = initCollection("actionLogs", []string{"orgid", "type"}, false)
	if err != nil {
		return err
	}

	err = initCollection("hlcHack", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("userDataRequests", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("qrcodes", []string{"orgid", "purposeid"}, true)
	if err != nil {
		return err
	}

	err = initCollection("webhooks", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("webhookDeliveries", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("policies", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("revisions", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("dataAgreements", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("dataAttributes", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("individuals", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("identityProviders", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("apiKeys", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("dataAgreementRecords", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("signatures", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("dataAgreementRecordsHistories", []string{"id"}, true)
	if err != nil {
		return err
	}

	return nil
}

func initCollection(collectionName string, keys []string, unique bool) error {

	c := DB.Client.Database(DB.Name).Collection(collectionName)

	indexOptions := options.Index()

	keysDoc := bson.D{}
	for _, key := range keys {
		keysDoc = append(keysDoc, bson.E{Key: key, Value: 1})
	}

	indexModel := mongo.IndexModel{
		Keys:    keysDoc,
		Options: indexOptions.SetSparse(true).SetUnique(unique),
	}

	_, err := c.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		log.Printf("error creating index on the specified keys: %v", err)
		return err
	}

	log.Printf("initialized %v collection", collectionName)
	return nil
}
