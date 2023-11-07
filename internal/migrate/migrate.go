package migrate

import (
	"context"
	"fmt"
	"time"

	"github.com/bb-consent/api/internal/apikey"
	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/org"
	"github.com/bb-consent/api/internal/policy"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Migrate() {
	migrateThirdPartyDataSharingToTrueInPolicyCollection()
	migrateThirdPartyDataSharingToTrueInDataAgreementsCollection()
	migrateNameInApiKeyCollection()
	migrateTimestampInDataAgreementsCollection()
	migrateTimestampInApiKeyCollection()
}

func migrateThirdPartyDataSharingToTrueInPolicyCollection() {
	policyCollection := policy.Collection()

	filter := bson.M{"thirdpartydatasharing": bson.M{"$nin": []interface{}{true, false}}}
	update := bson.M{"$set": bson.M{"thirdpartydatasharing": true}}
	_, err := policyCollection.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
	}
}

func migrateThirdPartyDataSharingToTrueInDataAgreementsCollection() {
	dataAgreementCollection := dataagreement.Collection()

	filter := bson.M{"policy.thirdpartydatasharing": bson.M{"$nin": []interface{}{true, false}}}
	update := bson.M{"$set": bson.M{"policy.thirdpartydatasharing": true}}

	_, err := dataAgreementCollection.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
	}
	// TODO: Handle impact towards revisions
}

func migrateNameInApiKeyCollection() {
	apiKeyCollection := apikey.Collection()

	filter := bson.M{"name": bson.M{"$exists": false}}
	update := bson.M{"$set": bson.M{"name": ""}}

	_, err := apiKeyCollection.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
	}
}

type dataAgreement struct {
	Id        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Timestamp string             `json:"timestamp"`
}

func migrateTimestampInDataAgreementsCollection() {
	dataAgreementCollection := dataagreement.Collection()

	// Get first organisation
	o, err := org.GetFirstOrganization()
	if err != nil {
		panic(err)
	}

	pipeline, err := dataagreement.CreatePipelineForFilteringDataAgreements(o.ID.Hex(), true)
	if err != nil {
		fmt.Println(err)
	}
	var results []dataAgreement
	cursor, err := dataAgreementCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &results); err != nil {
		fmt.Println(err)
	}
	var timestamp string

	for _, dataAgreement := range results {
		if len(dataAgreement.Timestamp) >= 1 {
			timestamp = dataAgreement.Timestamp
		} else {
			timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
		}

		filter := bson.M{"_id": dataAgreement.Id, "timestamp": bson.M{"$exists": false}}
		update := bson.M{"$set": bson.M{"timestamp": timestamp}}

		_, err := dataAgreementCollection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func migrateTimestampInApiKeyCollection() {
	apiKeyCollection := apikey.Collection()

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	filter := bson.M{"timestamp": bson.M{"$exists": false}}
	update := bson.M{"$set": bson.M{"timestamp": timestamp}}

	_, err := apiKeyCollection.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
	}
}
