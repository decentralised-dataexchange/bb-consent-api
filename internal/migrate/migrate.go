package migrate

import (
	"context"
	"fmt"
	"time"

	"github.com/bb-consent/api/internal/actionlog"
	"github.com/bb-consent/api/internal/apikey"
	"github.com/bb-consent/api/internal/dataagreement"
	dataagreementrecord "github.com/bb-consent/api/internal/dataagreement_record"
	dataagreementrecordhistory "github.com/bb-consent/api/internal/dataagreement_record_history"
	"github.com/bb-consent/api/internal/idp"
	"github.com/bb-consent/api/internal/image"
	"github.com/bb-consent/api/internal/individual"
	"github.com/bb-consent/api/internal/org"
	"github.com/bb-consent/api/internal/orgtype"
	"github.com/bb-consent/api/internal/otp"
	"github.com/bb-consent/api/internal/policy"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/signature"
	"github.com/bb-consent/api/internal/user"
	"github.com/bb-consent/api/internal/webhook"
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
	migrateConsentRecordIdAndIndividualIdInConsentHistoryCollection()
	migrateIdToStringInUsersCollection()
	migrateIdToStringInPoliciesCollection()
	migrateIdToStringInOrgTypesCollection()
	migrateIdToStringInOrganisationCollection()
	migrateIdToStringInApiKeyCollection()
	migrateIdToStringInActionLogsCollection()
	migrateIdToStringInIdentityProvidersCollection()
	migrateIdToStringInImagesCollection()
	migrateIdToStringInIndividualsCollection()
	migrateIdToStringInWebhooksCollection()
	migrateIdToStringInWebhookDeliveriesCollection()
	migrateIdToStringInOtpsCollection()
	migrateIdToStringInConsentRecordHistoriesCollection()
	migrateIdToStringInConsentRecordsCollection()
	migrateIdToStringInDataAgreementsCollection()
	migrateIdToStringInSignaturesCollection()
	migrateIdToStringInRevisionsCollection()
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
		fmt.Println(err)
		return
	}

	pipeline, err := dataagreement.CreatePipelineForFilteringDataAgreements(o.ID, true)
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
		return
	}
	organizationId := organization.ID

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

func migrateConsentRecordIdAndIndividualIdInConsentHistoryCollection() {
	consentHistoryCollection := dataagreementrecordhistory.Collection()

	filter := bson.M{"consentrecordid": bson.M{"$exists": false}}

	_, err := consentHistoryCollection.DeleteMany(context.TODO(), filter)
	if err != nil {
		fmt.Println(err)
	}
}

func migrateIdToStringInApiKeyCollection() {
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

		apiKeyId, err := primitive.ObjectIDFromHex(apiKey.Id)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": apiKeyId}

		exists, err := apiKeyCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = apiKeyCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := apiKeyCollection.InsertOne(context.TODO(), apiKey)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInActionLogsCollection() {
	actionLogCollection := actionlog.Collection()

	var results []actionlog.ActionLog

	cursor, err := actionLogCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, actionLog := range results {

		actionLogId, err := primitive.ObjectIDFromHex(actionLog.ID)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": actionLogId}

		exists, err := actionLogCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = actionLogCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := actionLogCollection.InsertOne(context.TODO(), actionLog)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInIdentityProvidersCollection() {
	idpCollection := idp.Collection()

	var results []idp.IdentityProvider

	cursor, err := idpCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, idp := range results {

		idpId, err := primitive.ObjectIDFromHex(idp.Id)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": idpId}

		exists, err := idpCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = idpCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := idpCollection.InsertOne(context.TODO(), idp)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInImagesCollection() {
	imageCollection := image.Collection()

	var results []image.Image

	cursor, err := imageCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, image := range results {

		imageId, err := primitive.ObjectIDFromHex(image.ID)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": imageId}

		exists, err := imageCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = imageCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := imageCollection.InsertOne(context.TODO(), image)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInIndividualsCollection() {
	individualCollection := individual.Collection()

	var results []individual.Individual

	cursor, err := individualCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, individual := range results {

		individualId, err := primitive.ObjectIDFromHex(individual.Id)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": individualId}

		exists, err := individualCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = individualCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := individualCollection.InsertOne(context.TODO(), individual)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInWebhooksCollection() {
	webhookCollection := webhook.WebhookCollection()

	var results []webhook.Webhook

	cursor, err := webhookCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, webhook := range results {

		webhookId, err := primitive.ObjectIDFromHex(webhook.ID)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": webhookId}

		exists, err := webhookCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = webhookCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := webhookCollection.InsertOne(context.TODO(), webhook)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInWebhookDeliveriesCollection() {
	webhookDeliveryCollection := webhook.WebhookDeliveryCollection()

	var results []webhook.WebhookDelivery

	cursor, err := webhookDeliveryCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, webhookDelivery := range results {

		webhookDeliveryId, err := primitive.ObjectIDFromHex(webhookDelivery.ID)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": webhookDeliveryId}

		exists, err := webhookDeliveryCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = webhookDeliveryCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := webhookDeliveryCollection.InsertOne(context.TODO(), webhookDelivery)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInOtpsCollection() {
	otpCollection := otp.Collection()

	var results []otp.Otp

	cursor, err := otpCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, otp := range results {

		otpId, err := primitive.ObjectIDFromHex(otp.ID)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": otpId}

		exists, err := otpCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = otpCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := otpCollection.InsertOne(context.TODO(), otp)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInPoliciesCollection() {
	policyCollection := policy.Collection()

	var results []policy.Policy

	cursor, err := policyCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, policy := range results {

		policyId, err := primitive.ObjectIDFromHex(policy.Id)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": policyId}

		exists, err := policyCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = policyCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := policyCollection.InsertOne(context.TODO(), policy)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInConsentRecordHistoriesCollection() {
	consentRecordHistoriesCollection := dataagreementrecordhistory.Collection()

	var results []dataagreementrecordhistory.DataAgreementRecordsHistory

	cursor, err := consentRecordHistoriesCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, consentRecordHistory := range results {

		consentRecordHistoryId, err := primitive.ObjectIDFromHex(consentRecordHistory.Id)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": consentRecordHistoryId}

		exists, err := consentRecordHistoriesCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = consentRecordHistoriesCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := consentRecordHistoriesCollection.InsertOne(context.TODO(), consentRecordHistory)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInConsentRecordsCollection() {
	consentRecordCollection := dataagreementrecord.Collection()

	var results []dataagreementrecord.DataAgreementRecord

	cursor, err := consentRecordCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, consentRecord := range results {

		consentRecordId, err := primitive.ObjectIDFromHex(consentRecord.Id)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": consentRecordId}

		exists, err := consentRecordCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = consentRecordCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := consentRecordCollection.InsertOne(context.TODO(), consentRecord)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInDataAgreementsCollection() {
	dataAgreementCollection := dataagreement.Collection()

	var results []dataagreement.DataAgreement

	cursor, err := dataAgreementCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, dataAgreement := range results {

		dataAgreementId, err := primitive.ObjectIDFromHex(dataAgreement.Id)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": dataAgreementId}

		exists, err := dataAgreementCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = dataAgreementCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := dataAgreementCollection.InsertOne(context.TODO(), dataAgreement)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInSignaturesCollection() {
	signatureCollection := signature.Collection()

	var results []signature.Signature

	cursor, err := signatureCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, signature := range results {

		signatureId, err := primitive.ObjectIDFromHex(signature.Id)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": signatureId}

		exists, err := signatureCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = signatureCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := signatureCollection.InsertOne(context.TODO(), signature)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInRevisionsCollection() {
	revisionCollection := revision.Collection()

	var results []revision.Revision

	cursor, err := revisionCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, revision := range results {

		revisionId, err := primitive.ObjectIDFromHex(revision.Id)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": revisionId}

		exists, err := revisionCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = revisionCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := revisionCollection.InsertOne(context.TODO(), revision)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInUsersCollection() {
	userCollection := user.Collection()

	var results []user.User

	cursor, err := userCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, user := range results {

		userId, err := primitive.ObjectIDFromHex(user.ID)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": userId}

		exists, err := userCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = userCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := userCollection.InsertOne(context.TODO(), user)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInOrgTypesCollection() {
	orgTypeCollection := orgtype.Collection()

	var results []orgtype.OrgType

	cursor, err := orgTypeCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, orgType := range results {

		orgTypeId, err := primitive.ObjectIDFromHex(orgType.ID)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": orgTypeId}

		exists, err := orgTypeCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = orgTypeCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := orgTypeCollection.InsertOne(context.TODO(), orgType)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}

func migrateIdToStringInOrganisationCollection() {
	orgCollection := org.Collection()

	var results []org.Organization

	cursor, err := orgCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println(err)
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		fmt.Println(err)
	}

	for _, org := range results {

		orgId, err := primitive.ObjectIDFromHex(org.ID)
		if err != nil {
			fmt.Println(err)
		}

		filter := bson.M{"_id": orgId}

		exists, err := orgCollection.CountDocuments(context.TODO(), filter)
		if err != nil {
			fmt.Println(err)
		}

		if exists > 0 {

			_, err = orgCollection.DeleteOne(context.TODO(), filter)
			if err != nil {
				fmt.Println(err)
			}

			_, err := orgCollection.InsertOne(context.TODO(), org)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

}
