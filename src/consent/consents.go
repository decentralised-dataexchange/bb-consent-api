package consent

import (
	"context"
	"time"

	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type consentStatus struct {
	Consented string
	TimeStamp time.Time
	Days      int
}

// Consent data type
type Consent struct {
	Status     consentStatus
	Value      string //Description??
	TemplateID string
}

// Purpose data type
type Purpose struct {
	ID       string `bson:"id,omitempty"`
	AllowAll bool
	Consents []Consent
}

// Consents data type
type Consents struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	OrgID    string
	UserID   string
	Purposes []Purpose
}

func collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("consents")
}

// Add Adds an consent to the collection
func Add(consent Consents) (Consents, error) {

	consent.ID = primitive.NewObjectID()
	_, err := collection().InsertOne(context.TODO(), &consent)
	return consent, err
}

// DeleteByUserOrg Deletes the consent by userID, orgID
func DeleteByUserOrg(userID string, orgID string) error {

	_, err := collection().DeleteMany(context.TODO(), bson.M{"userid": userID, "orgid": orgID})
	return err
}

// GetByUserOrg Get all consents of a user in organization
func GetByUserOrg(userID string, orgID string) (Consents, error) {

	var consents Consents
	err := collection().FindOne(context.TODO(), bson.M{"userid": userID, "orgid": orgID}).Decode(&consents)

	return consents, err
}

// Get Get consent by consentID
func Get(consentID string) (Consents, error) {
	var result Consents

	consentId, err := primitive.ObjectIDFromHex(consentID)
	if err != nil {
		return result, err
	}
	err = collection().FindOne(context.TODO(), bson.M{"_id": consentId}).Decode(&result)

	return result, err
}

// GetConsentedUsers Get list of users who are consented to an attribute
func GetConsentedUsers(orgID string, purposeID string, attributeID string, startID string, limit int) (userIDs []string, lastID string, err error) {
	c := collection()

	limit = 10000
	var results []Consents
	var cur *mongo.Cursor

	if startID == "" {
		pipeline := []bson.M{
			{"$match": bson.M{"orgid": orgID}},
			{"$unwind": "$purposes"},
			{"$unwind": "$purposes.consents"},
			{"$match": bson.M{
				"purposes.id":                        purposeID,
				"purposes.consents.templateid":       attributeID,
				"purposes.consents.status.consented": bson.M{"$regex": "^A"}},
			},
			{"$limit": int64(limit)},
		}
		cur, err = c.Aggregate(context.TODO(), pipeline)
	} else {
		pipeline := []bson.M{
			{"$match": bson.M{"orgid": orgID}},
			{"$unwind": "$purposes"},
			{"$unwind": "$purposes.consents"},
			{"$match": bson.M{
				"purposes.id":                        purposeID,
				"purposes.consents.templateid":       attributeID,
				"purposes.consents.status.consented": bson.M{"$regex": "^A"}},
			},
			{"$limit": int64(limit)},
			{"$gt": startID},
		}
		cur, err = c.Aggregate(context.TODO(), pipeline)
	}
	if err != nil {
		return
	}

	defer cur.Close(context.TODO())

	if err = cur.All(context.TODO(), &results); err != nil {
		return
	}

	for _, item := range results {
		userIDs = append(userIDs, item.UserID)
	}

	if len(results) != 0 && len(results) == (limit) {
		lastID = results[len(results)-1].ID.Hex()
	}

	return
}

// GetPurposeConsentedAllUsers Get all users with at-least one attribute consented in purpose.
func GetPurposeConsentedAllUsers(orgID string, purposeID string, startID string, limit int) (userIDs []string, lastID string, err error) {
	c := collection()

	limit = 10000
	var results []Consents
	var cur *mongo.Cursor

	if startID == "" {
		pipeline := []bson.M{
			{"$match": bson.M{"orgid": orgID}},
			{"$unwind": "$purposes"},
			{"$unwind": "$purposes.consents"},
			{"$match": bson.M{
				"purposes.id":                        purposeID,
				"purposes.consents.status.consented": bson.M{"$regex": "^A"}},
			},
			{"$limit": int64(limit)},
		}
		cur, err = c.Aggregate(context.TODO(), pipeline)
	} else {
		pipeline := []bson.M{
			{"$match": bson.M{"orgid": orgID}},
			{"$unwind": "$purposes"},
			{"$unwind": "$purposes.consents"},
			{"$match": bson.M{
				"purposes.id":                        purposeID,
				"purposes.consents.status.consented": bson.M{"$regex": "^A"}},
			},
			{"$limit": int64(limit)},
			{"$gt": startID},
		}
		cur, err = c.Aggregate(context.TODO(), pipeline)
	}
	if err != nil {
		return
	}

	defer cur.Close(context.TODO())

	if err = cur.All(context.TODO(), &results); err != nil {
		return
	}

	keys := make(map[string]bool)
	for _, item := range results {
		if _, value := keys[item.UserID]; !value {
			keys[item.UserID] = true
			userIDs = append(userIDs, item.UserID)
		}
	}

	if len(results) != 0 && len(results) == (limit) {
		lastID = results[len(results)-1].ID.Hex()
	}

	return
}

// UpdatePurposes Update consents purposes
func UpdatePurposes(consents Consents) (Consents, error) {
	c := collection()

	_, err := c.UpdateOne(context.TODO(), bson.M{"_id": consents.ID}, bson.M{"$set": bson.M{"purposes": consents.Purposes}})

	return consents, err
}
