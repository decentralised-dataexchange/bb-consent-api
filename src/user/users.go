package user

import (
	"context"
	"log"
	"time"

	"github.com/bb-consent/api/src/database"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/orgtype"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Org Organization snippet stored as part of user
type Org struct {
	OrgID        primitive.ObjectID `bson:"orgid,omitempty"`
	Name         string
	Location     string
	Type         string
	TypeID       primitive.ObjectID `bson:"typeid,omitempty"`
	EulaAccepted bool
}

// ClientInfo The client device details.
type ClientInfo struct {
	Token string
	Type  int
}

// Role Role assignment to user
type Role struct {
	RoleID int
	OrgID  string
}

// User data type
type User struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	Name              string
	IamID             string
	Email             string
	Phone             string
	ImageID           string
	ImageURL          string
	LastVisit         string //TODO Replace with ISODate()
	Client            ClientInfo
	Orgs              []Org
	APIKey            string
	Roles             []Role
	IncompleteProfile bool
}

func collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("users")
}

// Add Adds an user to the collection
func Add(user User) (User, error) {

	user.ID = primitive.NewObjectID()
	user.LastVisit = time.Now().Format(time.RFC3339)

	_, err := collection().InsertOne(context.TODO(), &user)

	return user, err
}

// Update Update the user details
func Update(userID string, u User) (User, error) {
	userId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return User{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": userId}, bson.M{"$set": u})
	if err != nil {
		return User{}, err
	}

	u, err = Get(userID)
	return u, err
}

// Delete Deletes the user by ID
func Delete(userID string) error {
	userId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": userId}

	_, err = collection().DeleteOne(context.TODO(), filter)

	return err
}

// GetByIamID Get the user by IamID
func GetByIamID(iamID string) (User, error) {
	var result User

	err := collection().FindOne(context.TODO(), bson.M{"iamid": iamID}).Decode(&result)
	if err != nil {
		log.Printf("Failed to find user id:%v err:%v", iamID, err)
		return result, err
	}

	return result, err
}

// Get Gets a single user by given id
func Get(userID string) (User, error) {
	c := collection()

	userId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return User{}, err
	}

	var result User

	// Find the user by ID
	filter := bson.M{"_id": userId}
	err = c.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		log.Printf("Failed to find user ID: %v, error: %v", userID, err)
		return result, err
	}

	// Update the last visited field
	t := time.Now().Format(time.RFC3339)
	update := bson.M{"$set": bson.M{"lastvisit": t}}
	updateOptions := options.Update().SetUpsert(false)
	_, err = c.UpdateOne(context.TODO(), filter, update, updateOptions)
	if err != nil {
		log.Printf("Failed to update LastVisit field for id:%v \n", userID)
	}

	return result, err
}

// GetByEmail Get user details by email
func GetByEmail(email string) (User, error) {
	var u User

	filter := bson.M{"email": email}

	projection := bson.M{"iamid": 1, "name": 1, "roles": 1}

	findOptions := options.FindOne().SetProjection(projection)

	err := collection().FindOne(context.TODO(), filter, findOptions).Decode(&u)

	return u, err
}

// EmailExist Check if email id is already in the collection
func EmailExist(email string) (bool, error) {
	filter := bson.M{"email": email}

	countOptions := options.Count().SetLimit(1)

	count, err := collection().CountDocuments(context.TODO(), filter, countOptions)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// PhoneNumberExist Check if phone number is already in the collection
func PhoneNumberExist(phone string) (bool, error) {
	filter := bson.M{"phone": phone}

	countOptions := options.Count().SetLimit(1)

	count, err := collection().CountDocuments(context.TODO(), filter, countOptions)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// UpdateClientDeviceInfo Update the client device info
func UpdateClientDeviceInfo(userID string, client ClientInfo) (User, error) {
	userId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return User{}, err
	}

	filter := bson.M{"_id": userId}
	update := bson.M{"$set": bson.M{"client": client}}

	_, err = collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return User{}, err
	}

	//TODO: Is this DB get necessary?
	u, err := Get(userID)
	return u, err
}

// AddRole Add roles to users
func AddRole(userID string, role Role) (User, error) {
	userId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return User{}, err
	}

	_, err = collection().UpdateOne(context.TODO(), bson.M{"_id": userId}, bson.M{"$push": bson.M{"roles": role}})
	if err != nil {
		return User{}, err
	}
	u, err := Get(userID)
	return u, err
}

// GetOrgSubscribeUsers Get list of users subscribed to an organizations
func GetOrgSubscribeUsers(orgID string, startID string, limit int) ([]User, string, error) {
	var results []User
	var err error
	limit = 10000

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "_id", Value: -1}})
	findOptions.SetLimit(int64(limit))

	orgId, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return results, "", err
	}

	filter := bson.M{"orgs.orgid": orgId}
	if startID != "" {
		startId, err := primitive.ObjectIDFromHex(startID)
		if err != nil {
			return results, "", err
		}

		filter["_id"] = bson.M{"$lt": startId}
	}

	cursor, err := collection().Find(context.TODO(), filter, findOptions)
	if err != nil {
		return results, "", err
	}
	defer cursor.Close(context.TODO())

	err = cursor.All(context.TODO(), &results)

	lastID := ""
	if err == nil {
		if len(results) != 0 && len(results) == (limit) {
			lastID = results[len(results)-1].ID.Hex()
		}
	}

	return results, lastID, nil

}

// GetOrgSubscribeIter Get Iterator to users subscribed to an organizations
func GetOrgSubscribeIter(orgID string) (*mongo.Cursor, error) {
	orgId, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"orgs.orgid": orgId}
	cursor, err := collection().Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	return cursor, nil
}

// GetOrgSubscribeCount Get count of users subscribed to an organizations
func GetOrgSubscribeCount(orgID string) (int64, error) {
	orgId, err := primitive.ObjectIDFromHex(orgID)
	if err != nil {
		return 0, err
	}
	filter := bson.M{"orgs.orgid": orgId}

	count, err := collection().CountDocuments(context.TODO(), filter)
	if err != nil {
		log.Printf("Failed to find user count by org id:%v err:%v", orgID, err)
		return 0, err
	}

	return count, err
}

// UpdateOrgTypeOfSubscribedUsers Updates the embedded organization type snippet for all users
func UpdateOrgTypeOfSubscribedUsers(orgType orgtype.OrgType) error {
	filter := bson.M{"orgs.typeid": orgType.ID}

	cursor, err := collection().Find(context.TODO(), filter)
	if err != nil {
		return err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var u User
		err := cursor.Decode(&u)
		if err != nil {
			return err
		}

		for i := range u.Orgs {
			if u.Orgs[i].TypeID == orgType.ID {
				u.Orgs[i].Type = orgType.Type
			}
		}

		_, err = collection().ReplaceOne(context.TODO(), bson.M{"_id": u.ID}, u)
		if err != nil {
			return err
		}
	}

	log.Println("Successfully updated users for organization type name change")
	return nil
}

// UpdateOrganizationsSubscribedUsers Updates the embedded organization snippet for all users
func UpdateOrganizationsSubscribedUsers(org org.Organization) error {
	filter := bson.M{"orgs.orgid": org.ID}
	cursor, err := collection().Find(context.TODO(), filter)
	if err != nil {
		return err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var result User
		err := cursor.Decode(&result)
		if err != nil {
			return err
		}

		for i := range result.Orgs {
			if result.Orgs[i].OrgID == org.ID {
				result.Orgs[i].Name = org.Name
				result.Orgs[i].Location = org.Location
			}
		}

		_, err = collection().ReplaceOne(context.TODO(), bson.M{"_id": result.ID}, result)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateOrganization Updates organization to user collection
func UpdateOrganization(userID string, org Org) (User, error) {
	var result User

	userId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return User{}, err
	}

	filter := bson.M{"_id": userId}

	update := bson.M{"$push": bson.M{"orgs": org}}

	_, err = collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return result, err
	}

	err = collection().FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return result, err
	}

	return result, err
}

func GetUserOrgDetails(u User, oID string) (org Org, found bool) {
	for _, o := range u.Orgs {
		if o.OrgID.Hex() == oID {
			return o, true
		}
	}
	return org, false
}

// DeleteOrganization Remove user from an organization
func DeleteOrganization(userID string, orgID string) (User, error) {

	u, err := Get(userID)
	if err != nil {
		return User{}, err
	}
	org, _ := GetUserOrgDetails(u, orgID)
	//Check found == true

	userId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return User{}, err
	}

	filter := bson.M{"_id": userId}
	update := bson.M{"$pull": bson.M{"orgs": org}}

	var result User

	_, err = collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return result, err
	}

	err = collection().FindOne(context.TODO(), filter).Decode(&result)

	return result, err
}

// RemoveRole Remove role of an user
func RemoveRole(userID string, role Role) (User, error) {
	userId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return User{}, err
	}
	filter := bson.M{"_id": userId}
	update := bson.M{"$pull": bson.M{"roles": role}}

	_, err = collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return User{}, err
	}

	u, err := Get(userID)
	return u, err
}

// UpdateAPIKey update apikey to user
func UpdateAPIKey(userID string, apiKey string) error {
	userId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": userId}
	update := bson.M{"$set": bson.M{"apikey": apiKey}}

	_, err = collection().UpdateOne(context.TODO(), filter, update)
	return err
}

// GetAPIKey Gets the API key of the user
func GetAPIKey(userID string) (string, error) {
	userId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "", err
	}

	projection := bson.M{"apikey": 1}
	opts := options.FindOne().SetProjection(projection)

	var result User
	err = collection().FindOne(context.TODO(), bson.M{"_id": userId}, opts).Decode(&result)
	if err != nil {
		log.Printf("Failed to find user by id:%v err:%v", userID, err)
		return "", err
	}

	return result.APIKey, err
}
