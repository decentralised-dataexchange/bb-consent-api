package orgtype

import (
	"context"
	"log"

	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// OrgType Type related information
type OrgType struct {
	ID       string `bson:"_id,omitempty"`
	Type     string
	ImageID  string
	ImageURL string
}

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("orgTypes")
}

// Add Adds an organization
func Add(ot OrgType) (OrgType, error) {

	ot.ID = primitive.NewObjectID().Hex()
	_, err := Collection().InsertOne(context.TODO(), &ot)

	return ot, err
}

// Get Gets organization type by given id
func Get(organizationTypeID string) (OrgType, error) {
	var result OrgType

	err := Collection().FindOne(context.Background(), bson.M{"_id": organizationTypeID}).Decode(&result)

	return result, err
}

// GetAll Gets all organization types
func GetAll() ([]OrgType, error) {

	var results []OrgType

	cursor, err := Collection().Find(context.TODO(), bson.M{})
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

	_, err := Collection().UpdateOne(context.TODO(), bson.M{"_id": organizationTypeID}, bson.M{"$set": bson.M{"type": typeName}})
	if err == nil {
		return Get(organizationTypeID)
	}
	return OrgType{}, err
}

// Delete Deletes an organization
func Delete(organizationTypeID string) error {

	_, err := Collection().DeleteOne(context.TODO(), bson.M{"_id": organizationTypeID})

	return err
}

// UpdateImage Update the org type image
func UpdateImage(organizationTypeID string, imageID string, imageURL string) error {

	_, err := Collection().UpdateOne(context.TODO(), bson.M{"_id": organizationTypeID},
		bson.M{"$set": bson.M{"imageid": imageID, "imageurl": imageURL}})
	return err
}

// GetTypesCount Gets types count
func GetTypesCount() (int64, error) {
	count, err := Collection().CountDocuments(context.TODO(), bson.D{})
	if err != nil {
		return count, err
	}

	return count, err
}

// GetFirstOrganization Gets first type
func GetFirstType() (OrgType, error) {

	var result OrgType
	err := Collection().FindOne(context.TODO(), bson.M{}).Decode(&result)

	return result, err
}

// DeleteAllTypes delete all types
func DeleteAllTypes() (*mongo.DeleteResult, error) {

	result, err := Collection().DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		return result, err
	}
	log.Printf("Number of documents deleted: %d\n", result.DeletedCount)

	return result, err
}

// AddOrganizationType Adds an organization type
func AddOrganizationType(typeReq config.GlobalPolicy) (OrgType, error) {

	var orgType OrgType
	orgType.Type = typeReq.IndustrySector

	orgType, err := Add(orgType)
	if err != nil {
		log.Printf("Failed to add organization type: %v", orgType)
		return OrgType{}, err
	}
	return orgType, err
}
