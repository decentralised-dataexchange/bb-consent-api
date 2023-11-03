package migrate

import (
	"context"
	"fmt"

	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/policy"
	"go.mongodb.org/mongo-driver/bson"
)

func Migrate() {
	migrateThirdPartyDataSharingToTrueInPolicyCollection()
	migrateThirdPartyDataSharingToTrueInDataAgreementsCollection()
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