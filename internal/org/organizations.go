package org

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/database"
	"github.com/bb-consent/api/internal/orgtype"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Admin Users
type Admin struct {
	UserID string
	RoleID int
}

// Organization organization data type
type Organization struct {
	ID            string `bson:"_id,omitempty"`
	Name          string
	CoverImageID  string
	CoverImageURL string
	LogoImageID   string
	LogoImageURL  string
	Location      string
	Type          orgtype.OrgType
	Description   string
	Enabled       bool
	PolicyURL     string
	EulaURL       string
	Admins        []Admin
	Subs          Subscribe
}

// Subscribe Defines how users can subscribe to organization
type Subscribe struct {
	Method int
	Key    string
}

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("organizations")
}

// Add Adds an organization
func Add(org Organization) (Organization, error) {

	org.ID = primitive.NewObjectID().Hex()
	_, err := Collection().InsertOne(context.TODO(), &org)
	if err != nil {
		return org, err
	}
	return org, nil
}

// Get Gets a single organization by given id
func Get(organizationId string) (Organization, error) {

	var result Organization
	err := Collection().FindOne(context.TODO(), bson.M{"_id": organizationId}).Decode(&result)

	return result, err
}

// GetFirstOrganization Gets first organization
func GetFirstOrganization() (Organization, error) {

	var result Organization
	err := Collection().FindOne(context.TODO(), bson.M{}).Decode(&result)

	return result, err
}

// Update Updates the organization
func Update(org Organization) (Organization, error) {

	filter := bson.M{"_id": org.ID}
	update := bson.M{"$set": org}

	_, err := Collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return org, err
	}
	return org, err
}

// UpdateCoverImage Update the organization image
func UpdateCoverImage(organizationID string, imageID string) (Organization, error) {

	_, err := Collection().UpdateOne(context.TODO(), bson.M{"_id": organizationID}, bson.M{"$set": bson.M{"coverimageid": imageID}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// UpdateLogoImage Update the organization image
func UpdateLogoImage(organizationID string, imageID string) (Organization, error) {

	_, err := Collection().UpdateOne(context.TODO(), bson.M{"_id": organizationID}, bson.M{"$set": bson.M{"logoimageid": imageID}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// AddAdminUsers Add admin users to organization
func AddAdminUsers(organizationID string, admin Admin) (Organization, error) {

	_, err := Collection().UpdateOne(context.TODO(), bson.M{"_id": organizationID}, bson.M{"$push": bson.M{"admins": admin}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// GetAdminUsers Get admin users of organization
func GetAdminUsers(organizationID string) (Organization, error) {

	filter := bson.M{"_id": organizationID}
	projection := bson.M{"admins": 1}

	findOptions := options.FindOne().SetProjection(projection)

	var result Organization
	err := Collection().FindOne(context.TODO(), filter, findOptions).Decode(&result)

	return result, err
}

// DeleteAdminUsers Delete admin users from organization
func DeleteAdminUsers(organizationID string, admin Admin) (Organization, error) {

	_, err := Collection().UpdateOne(context.TODO(), bson.M{"_id": organizationID}, bson.M{"$pull": bson.M{"admins": admin}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// UpdateOrganizationsOrgType Updates the embedded organization type snippet of all Organization
func UpdateOrganizationsOrgType(oType orgtype.OrgType) error {

	filter := bson.M{"type._id": oType.ID}
	update := bson.M{"$set": bson.M{"type": oType}}

	_, err := Collection().UpdateMany(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	log.Println("successfully updated organiztions for type name change")
	return nil
}

// SetEnabled Sets the enabled status to true/false
func SetEnabled(organizationID string, enabled bool) (Organization, error) {

	_, err := Collection().UpdateOne(context.TODO(), bson.M{"_id": organizationID}, bson.M{"$set": bson.M{"enabled": enabled}})
	if err != nil {
		return Organization{}, err
	}
	o, err := Get(organizationID)
	return o, err
}

// GetSubscribeMethod Get org subscribe method
func GetSubscribeMethod(orgId string) (int, error) {
	var result Organization

	filter := bson.M{"_id": orgId}
	projection := bson.M{"subs.method": 1}

	findOptions := options.FindOne().SetProjection(projection)

	err := Collection().FindOne(context.TODO(), filter, findOptions).Decode(&result)

	return result.Subs.Method, err
}

// UpdateSubscribeMethod Update subscription method
func UpdateSubscribeMethod(orgId string, method int) error {

	filter := bson.M{"_id": orgId}
	update := bson.M{"$set": bson.M{"subs.method": method}}

	_, err := Collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// UpdateSubscribeKey Update subscription key
func UpdateSubscribeKey(orgId string, key string) error {

	filter := bson.M{"_id": orgId}
	update := bson.M{"$set": bson.M{"subs.key": key}}

	_, err := Collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// GetSubscribeKey Update subscription token
func GetSubscribeKey(orgId string) (string, error) {
	var result Organization

	filter := bson.M{"_id": orgId}
	projection := bson.M{"subs.key": 1}
	findOptions := options.FindOne().SetProjection(projection)

	err := Collection().FindOne(context.TODO(), filter, findOptions).Decode(&result)

	return result.Subs.Key, err
}

// GetName Get organization name by given id
func GetName(organizationID string) (string, error) {
	var result Organization

	filter := bson.M{"_id": organizationID}
	projection := bson.M{"name": 1}
	findOptions := options.FindOne().SetProjection(projection)

	err := Collection().FindOne(context.TODO(), filter, findOptions).Decode(&result)

	return result.Name, err
}

// GetOrganizationsCount Get organizations count
func GetOrganizationsCount() (int64, error) {
	count, err := Collection().CountDocuments(context.TODO(), bson.D{})
	if err != nil {
		return count, err
	}

	return count, err
}

// DeleteAllOrganizations delete all organizations
func DeleteAllOrganizations() (*mongo.DeleteResult, error) {

	result, err := Collection().DeleteMany(context.TODO(), bson.D{})
	if err != nil {
		return result, err
	}
	log.Printf("Number of documents deleted: %d\n", result.DeletedCount)

	return result, err
}

// AddOrganization Adds an organization
func AddOrganization(orgReq config.Organization, typeId string, userId string) (Organization, error) {

	// validating request payload
	valid, err := govalidator.ValidateStruct(orgReq)
	if !valid {
		log.Printf("Missing mandatory params for adding organization")
		return Organization{}, err
	}

	// checking if the string contained whitespace only
	if strings.TrimSpace(orgReq.Name) == "" {
		log.Printf("Failed to add organization: Missing mandatory param - Name")
		return Organization{}, errors.New("missing mandatory param - Name")
	}

	if strings.TrimSpace(orgReq.Location) == "" {
		log.Printf("Failed to add organization: Missing mandatory param - Location")
		return Organization{}, errors.New("missing mandatory param - Location")
	}

	description := orgReq.Description
	defaultDescription := "is committed to safeguarding your privacy. We process your personal data in line with data agreements, ensuring adherence to ISO27560 standards and legal frameworks like GDPR. For every personal data we process, you can view its usage purpose and make informed choices to opt in or out. For inquiries, contact our Data Protection Officer at dpo@"

	if strings.TrimSpace(description) == "" {
		description = defaultDescription
	}

	orgType, err := orgtype.Get(typeId)
	if err != nil {
		log.Printf("Invalid organization type ID: %v", typeId)
		return Organization{}, err
	}

	admin := Admin{UserID: userId, RoleID: common.GetRoleID("Admin")}

	var o Organization
	o.Name = orgReq.Name
	o.Location = orgReq.Location
	o.Type = orgType
	o.Description = description
	o.EulaURL = orgReq.EulaURL
	o.Admins = append(o.Admins, admin)
	o.Enabled = true

	orgResp, err := Add(o)
	if err != nil {
		log.Printf("Failed to add organization: %v", orgReq.Name)
		return orgResp, err
	}

	return orgResp, err
}
