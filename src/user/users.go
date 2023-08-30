package user

import (
	"log"
	"time"

	"github.com/bb-consent/api/src/database"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/orgtype"
	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Org Organization snippet stored as part of user
type Org struct {
	OrgID        bson.ObjectId `bson:"orgid,omitempty"`
	Name         string
	Location     string
	Type         string
	TypeID       bson.ObjectId `bson:"typeid,omitempty"`
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
	ID                bson.ObjectId `bson:"_id,omitempty"`
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

func session() *mgo.Session {
	return database.DB.Session.Copy()
}

func collection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("users")
}

// Add Adds an user to the collection
func Add(user User) (User, error) {
	s := session()
	defer s.Close()

	user.ID = bson.NewObjectId()
	user.LastVisit = time.Now().Format(time.RFC3339)

	return user, collection(s).Insert(&user)
}

// Update Update the user details
func Update(userID string, u User) (User, error) {
	s := session()
	defer s.Close()

	err := collection(s).UpdateId(bson.ObjectIdHex(userID), u)
	if err != nil {
		return User{}, err
	}
	u, err = Get(userID)
	return u, err
}

// Delete Deletes the user by ID
func Delete(userID string) error {
	s := session()
	defer s.Close()

	return collection(s).RemoveId(bson.ObjectIdHex(userID))
}

// GetByIamID Get the user by IamID
func GetByIamID(iamID string) (User, error) {
	var result User
	s := session()
	defer s.Close()

	err := collection(s).Find(bson.M{"iamid": iamID}).One(&result)

	if err != nil {
		log.Printf("Failed to find user id:%v err:%v", iamID, err)
		return result, err
	}

	return result, err
}

// Get Gets a single user by given id
func Get(userID string) (User, error) {
	s := session()
	defer s.Close()
	c := collection(s)

	var result User
	err := c.FindId(bson.ObjectIdHex(userID)).One(&result)

	if err != nil {
		log.Printf("Failed to find user id:%v err:%v", userID, err)
		return result, err
	}

	//Update the last visited field
	t := time.Now().Format(time.RFC3339)
	err = c.Update(bson.M{"_id": bson.ObjectIdHex(userID)}, bson.M{"$set": bson.M{"lastvisit": t}})
	if err != nil {
		log.Printf("Failed to update LastVisit field for id:%v \n", userID)
	}

	return result, err
}

// GetByEmail Get user details by email
func GetByEmail(email string) (User, error) {
	s := session()
	defer s.Close()

	var u User

	err := collection(s).Find(bson.M{"email": email}).Select(bson.M{"iamid": 1, "name": 1, "roles": 1}).One(&u)

	return u, err
}

// EmailExist Check if email id is already in the collection
func EmailExist(email string) (bool, error) {
	s := session()
	defer s.Close()

	q := collection(s).Find(bson.M{"email": email}).Limit(1)

	c, err := q.Count()
	if err != nil {
		return false, err
	}

	if c == 0 {
		return false, err
	}

	return true, err
}

// PhoneNumberExist Check if phone number is already in the collection
func PhoneNumberExist(phone string) (bool, error) {
	s := session()
	defer s.Close()

	q := collection(s).Find(bson.M{"phone": phone}).Limit(1)

	c, err := q.Count()
	if err != nil {
		return false, err
	}

	if c == 0 {
		return false, err
	}

	return true, err
}

// UpdateClientDeviceInfo Update the client device info
func UpdateClientDeviceInfo(userID string, client ClientInfo) (User, error) {
	s := session()
	defer s.Close()
	c := collection(s)

	err := c.Update(bson.M{"_id": bson.ObjectIdHex(userID)}, bson.M{"$set": bson.M{"client": client}})
	if err != nil {
		return User{}, err
	}
	//TODO: Is this DB get necessary?
	u, err := Get(userID)
	return u, err
}

// AddRole Add roles to users
func AddRole(userID string, role Role) (User, error) {
	s := session()
	defer s.Close()

	err := collection(s).Update(bson.M{"_id": bson.ObjectIdHex(userID)}, bson.M{"$push": bson.M{"roles": role}})
	if err != nil {
		return User{}, err
	}
	u, err := Get(userID)
	return u, err
}

// GetOrgSubscribeUsers Get list of users subscribed to an organizations
func GetOrgSubscribeUsers(orgID string, startID string, limit int) ([]User, string, error) {
	s := session()
	defer s.Close()

	var results []User
	var err error
	limit = 10000
	if startID == "" {
		err = collection(s).Find(bson.M{"orgs.orgid": bson.ObjectIdHex(orgID)}).Select(bson.M{"name": 1, "phone": 1, "email": 1}).Sort("-_id").Limit(limit).All(&results)
	} else {
		err = collection(s).Find(bson.M{"orgs.orgid": bson.ObjectIdHex(orgID), "_id": bson.M{"$lt": bson.ObjectIdHex(startID)}}).Select(bson.M{"name": 1, "phone": 1, "email": 1}).Sort("-_id").Limit(limit).All(&results)
	}

	var lastID = ""
	if err == nil {
		if len(results) != 0 && len(results) == (limit) {
			lastID = results[len(results)-1].ID.Hex()
		}
	}

	return results, lastID, err
}

// GetOrgSubscribeIter Get Iterator to users subscribed to an organizations
func GetOrgSubscribeIter(orgID string) *mgo.Iter {
	s := session()
	defer s.Close()

	iter := collection(s).Find(bson.M{"orgs.orgid": bson.ObjectIdHex(orgID)}).Iter()

	return iter
}

// GetOrgSubscribeCount Get count of users subscribed to an organizations
func GetOrgSubscribeCount(orgID string) (int, error) {
	s := session()
	defer s.Close()

	count, err := collection(s).Find(bson.M{"orgs.orgid": bson.ObjectIdHex(orgID)}).Count()

	if err != nil {
		log.Printf("Failed to find user count by org id:%v err:%v", orgID, err)
		return 0, err
	}

	return count, err
}

// UpdateOrgTypeOfSubscribedUsers Updates the embedded organization type snippet for all users
func UpdateOrgTypeOfSubscribedUsers(orgType orgtype.OrgType) error {
	s := session()
	defer s.Close()
	c := collection(s)

	var u User
	iter := c.Find(bson.M{"orgs.typeid": orgType.ID}).Iter()
	for iter.Next(&u) {
		for i := range u.Orgs {
			if u.Orgs[i].TypeID == orgType.ID {
				u.Orgs[i].Type = orgType.Type
			}
			err := c.UpdateId(u.ID, u)
			if err != nil {
				return err
			}
		}
	}
	if err := iter.Close(); err != nil {
		return err
	}
	log.Println("successfully updated users for organization type name change")
	return nil
}

// UpdateOrganizationsSubscribedUsers Updates the embedded organization snippet for all users
func UpdateOrganizationsSubscribedUsers(org org.Organization) error {
	s := session()
	defer s.Close()
	c := collection(s)

	var result User
	iter := c.Find(bson.M{"orgs.orgid": org.ID}).Iter()
	for iter.Next(&result) {
		for i := range result.Orgs {
			if result.Orgs[i].OrgID == org.ID {
				result.Orgs[i].Name = org.Name
				result.Orgs[i].Location = org.Location
			}
			err := c.UpdateId(result.ID, result)
			if err != nil {
				return err
			}
		}
	}
	if err := iter.Close(); err != nil {
		return err
	}
	return nil
}

// UpdateOrganization Updates organization to user collection
func UpdateOrganization(userID string, org Org) (User, error) {
	s := session()
	defer s.Close()
	c := collection(s)

	var result User
	err := c.Update(bson.M{"_id": bson.ObjectIdHex(userID)}, bson.M{"$push": bson.M{"orgs": org}})
	if err != nil {
		return result, err
	}

	err = c.FindId(bson.ObjectIdHex(userID)).One(&result)
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
	s := session()
	defer s.Close()
	c := collection(s)

	u, err := Get(userID)
	if err != nil {
		return User{}, err
	}
	org, _ := GetUserOrgDetails(u, orgID)
	//Check found == true

	var result User
	err = c.Update(bson.M{"_id": bson.ObjectIdHex(userID)}, bson.M{"$pull": bson.M{"orgs": org}})
	if err != nil {
		return result, err
	}

	err = c.FindId(bson.ObjectIdHex(userID)).One(&result)
	return result, err
}

// RemoveRole Remove role of an user
func RemoveRole(userID string, role Role) (User, error) {
	s := session()
	defer s.Close()

	err := collection(s).Update(bson.M{"_id": bson.ObjectIdHex(userID)}, bson.M{"$pull": bson.M{"roles": role}})
	if err != nil {
		return User{}, err
	}
	u, err := Get(userID)
	return u, err
}

// UpdateAPIKey update apikey to user
func UpdateAPIKey(userID string, apiKey string) error {
	s := session()
	defer s.Close()
	c := collection(s)

	err := c.Update(bson.M{"_id": bson.ObjectIdHex(userID)}, bson.M{"$set": bson.M{"apikey": apiKey}})
	return err
}

// GetAPIKey Gets the API key of the user
func GetAPIKey(userID string) (string, error) {
	s := session()
	defer s.Close()

	var result User
	err := collection(s).FindId(bson.ObjectIdHex(userID)).Select(bson.M{"apikey": 1}).One(&result)

	if err != nil {
		log.Printf("Failed to find user by id:%v err:%v", userID, err)
		return "", err
	}

	return result.APIKey, err
}
