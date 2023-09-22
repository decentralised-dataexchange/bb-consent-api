package orgtype

import (
	"context"

	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// OrgType Type related information
type OrgType struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Type     string
	ImageID  string
	ImageURL string
}

func collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("orgTypes")
}

// Add Adds an organization
func Add(ot OrgType) (OrgType, error) {

	ot.ID = primitive.NewObjectID()
	_, err := collection().InsertOne(context.TODO(), &ot)

	return ot, err
}

// Get Gets organization type by given id
func Get(organizationTypeID string) (OrgType, error) {
	var result OrgType

	orgTypeID, err := primitive.ObjectIDFromHex(organizationTypeID)
	if err != nil {
		return result, err
	}

	err = collection().FindOne(context.Background(), bson.M{"_id": orgTypeID}).Decode(&result)

	return result, err
}

// GetAll Gets all organization types
func GetAll() ([]OrgType, error) {

	var results []OrgType

	cursor, err := collection().Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, err
}

// Update Update the organization type
func Update(organizationTypeID string, typeName string) (OrgType, error) {
	orgTypeID, err := primitive.ObjectIDFromHex(organizationTypeID)
	if err != nil {
		return OrgType{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgTypeID}, bson.M{"$set": bson.M{"type": typeName}})
	if err == nil {
		return Get(organizationTypeID)
	}
	return OrgType{}, err
}

// Delete Deletes an organization
func Delete(organizationTypeID string) error {
	orgTypeID, err := primitive.ObjectIDFromHex(organizationTypeID)
	if err != nil {
		return err
	}

	_, err = collection().DeleteOne(context.TODO(), bson.M{"_id": orgTypeID})

	return err
}

// UpdateImage Update the org type image
func UpdateImage(organizationTypeID string, imageID string, imageURL string) error {
	orgTypeID, err := primitive.ObjectIDFromHex(organizationTypeID)
	if err != nil {
		return err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": orgTypeID},
		bson.M{"$set": bson.M{"imageid": imageID, "imageurl": imageURL}})
	return err
}
