package migrate

import (
	"context"
	"fmt"
	"time"

	"github.com/bb-consent/api/internal/apikey"
	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/idp"
	"github.com/bb-consent/api/internal/individual"
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
	migrateOrganisationIdInIDPCollection()
	migrateExpiryTimestampInApiKeyCollection()
	migrateUnusedFieldsFromOrganistaionColloction()
	migrateIsOnboardedFromIDPInindividualCollection()
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

func migrateOrganisationIdInIDPCollection() {
	organization, err := org.GetFirstOrganization()
	if err != nil {
		fmt.Println(err)
	}
	organizationId := organization.ID.Hex()

	idpCollection := idp.Collection()

	filter := bson.M{"organisationid": ""}
	update := bson.M{"$set": bson.M{"organisationid": organizationId}}

	_, err = idpCollection.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
	}
}

func migrateExpiryTimestampInApiKeyCollection() {
	apiKeyCollection := apikey.Collection()

	var results []apikey.ApiKey

	cursor, err := apiKeyCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, apiKey := range results {
		createTimestamp := apiKey.Timestamp
		expiryInDays := apiKey.ExpiryInDays
		// Parse the timestamp
		creationTime, err := time.Parse(time.RFC3339, createTimestamp)
		if err != nil {
			fmt.Println(err)
		}
		expiryTime := creationTime.Add(time.Duration(24*expiryInDays) * time.Hour)

		expiryTimestamp := expiryTime.UTC().Format("2006-01-02T15:04:05Z")

		filter := bson.M{"_id": apiKey.Id, "expirytimestamp": bson.M{"$exists": false}}
		update := bson.M{"$set": bson.M{"expirytimestamp": expiryTimestamp}}

		_, err = apiKeyCollection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func migrateUnusedFieldsFromOrganistaionColloction() {

	orgCollection := org.Collection()

	filter := bson.M{}
	update := bson.M{"$unset": bson.M{"jurisdiction": 1,
		"disclosure":                        1,
		"restriction":                       1,
		"shared3pp":                         1,
		"templates":                         1,
		"purposes":                          1,
		"hlcsupport":                        1,
		"dataretention":                     1,
		"identityproviderrepresentation":    1,
		"keycloakopenidclient":              1,
		"externalidentityprovideravailable": 1,
	}}

	_, err := orgCollection.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
	}
}

func migrateIsOnboardedFromIDPInindividualCollection() {
	individualCollection := individual.Collection()

	filter := bson.M{}
	update := bson.M{
		"$rename": bson.M{"isonboardedfromid": "isonboardedfromidp"},
	}

	_, err := individualCollection.UpdateMany(context.TODO(), filter, update)
	if err != nil {
		fmt.Println(err)
	}
}
