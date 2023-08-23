package user

import (
	"log"
	"time"

	"github.com/bb-consent/api/src/database"
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
