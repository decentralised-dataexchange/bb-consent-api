package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/database"
	"github.com/bb-consent/api/internal/org"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Org Organization snippet stored as part of user
type Org struct {
	OrgID        string `bson:"orgid,omitempty" json:"id"`
	Name         string `json:"name"`
	Location     string `json:"location"`
	Type         string `json:"type"`
	TypeID       string `bson:"typeid,omitempty" json:"typeId"`
	EulaAccepted bool   `json:"eulaAccepted"`
}

// Role Role assignment to user
type Role struct {
	RoleID int    `json:"roleId"`
	OrgID  string `json:"orgId"`
}

// User data type
type User struct {
	ID                 string `bson:"_id,omitempty" json:"id"`
	Name               string `json:"name"`
	ExternalId         string `json:"externalId"`
	ExternalIdType     string `json:"externalIdType"`
	IdentityProviderId string `json:"identityProviderId"`
	IamID              string `json:"iamId"`
	Email              string `json:"email"`
	Phone              string `json:"phone"`
	ImageID            string `json:"imageId"`
	ImageURL           string `json:"imageUrl"`
	LastVisit          string `json:"lastVisit"` //TODO Replace with ISODate()
	Orgs               []Org  `json:"orgs"`
	APIKey             string `json:"apiKey"`
	Roles              []Role `json:"roles"`
	IncompleteProfile  bool   `json:"incompleteProfile"`
}

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("users")
}

// Add Adds an user to the collection
func Add(user User) (User, error) {

	user.ID = primitive.NewObjectID().Hex()
	user.LastVisit = time.Now().Format(time.RFC3339)

	_, err := Collection().InsertOne(context.TODO(), &user)

	return user, err
}

// Update Update the user details
func Update(userId string, u User) (User, error) {

	_, err := Collection().UpdateOne(context.TODO(), bson.M{"_id": userId}, bson.M{"$set": u})
	if err != nil {
		return User{}, err
	}

	u, err = Get(userId)
	return u, err
}

// Delete Deletes the user by ID
func Delete(userId string) error {

	filter := bson.M{"_id": userId}

	_, err := Collection().DeleteOne(context.TODO(), filter)

	return err
}

// GetByIamID Get the user by IamID
func GetByIamID(iamID string) (User, error) {
	var result User

	err := Collection().FindOne(context.TODO(), bson.M{"iamid": iamID}).Decode(&result)
	if err != nil {
		log.Printf("Failed to find user id:%v err:%v", iamID, err)
		return result, err
	}

	return result, err
}

// Get Gets a single user by given id
func Get(userId string) (User, error) {
	c := Collection()

	var result User

	// Find the user by ID
	filter := bson.M{"_id": userId}
	err := c.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		log.Printf("Failed to find user ID: %v, error: %v", userId, err)
		return result, err
	}

	// Update the last visited field
	t := time.Now().Format(time.RFC3339)
	update := bson.M{"$set": bson.M{"lastvisit": t}}
	updateOptions := options.Update().SetUpsert(false)
	_, err = c.UpdateOne(context.TODO(), filter, update, updateOptions)
	if err != nil {
		log.Printf("Failed to update LastVisit field for id:%v", userId)
	}

	return result, err
}

// GetByEmail Get user details by email
func GetByEmail(email string) (User, error) {
	var u User

	filter := bson.M{"email": email}

	projection := bson.M{"iamid": 1, "name": 1, "roles": 1}

	findOptions := options.FindOne().SetProjection(projection)

	err := Collection().FindOne(context.TODO(), filter, findOptions).Decode(&u)

	return u, err
}

// EmailExist Check if email id is already in the collection
func EmailExist(email string) (bool, error) {
	filter := bson.M{"email": email}

	countOptions := options.Count().SetLimit(1)

	count, err := Collection().CountDocuments(context.TODO(), filter, countOptions)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// PhoneNumberExist Check if phone number is already in the collection
func PhoneNumberExist(phone string) (bool, error) {
	filter := bson.M{"phone": phone}

	countOptions := options.Count().SetLimit(1)

	count, err := Collection().CountDocuments(context.TODO(), filter, countOptions)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// AddRole Add roles to users
func AddRole(userId string, role Role) (User, error) {

	_, err := Collection().UpdateOne(context.TODO(), bson.M{"_id": userId}, bson.M{"$push": bson.M{"roles": role}})
	if err != nil {
		return User{}, err
	}
	u, err := Get(userId)
	return u, err
}

// UpdateOrganizationsSubscribedUsers Updates the embedded organization snippet for all users
func UpdateOrganizationsSubscribedUsers(org org.Organization) error {
	filter := bson.M{"orgs.orgid": org.ID}
	cursor, err := Collection().Find(context.TODO(), filter)
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

		_, err = Collection().ReplaceOne(context.TODO(), bson.M{"_id": result.ID}, result)
		if err != nil {
			return err
		}
	}

	return nil
}

type iamCredentials struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type iamUserRegisterReq struct {
	Username    string           `json:"username"`
	Firstname   string           `json:"firstName"`
	Lastname    string           `json:"lastName"`
	Email       string           `json:"email"`
	Enabled     bool             `json:"enabled"`
	RealmRoles  []string         `json:"realmRoles"`
	Credentials []iamCredentials `json:"credentials"`
}

type iamToken struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
}

type iamError struct {
	ErrorType string `json:"error"`
	Error     string `json:"error_description"`
}

func getAdminToken(iamConfig config.Iam) (iamToken, int, iamError, error) {
	t, status, iamErr, err := getToken(iamConfig.AdminUser, iamConfig.AdminPassword, "admin-cli", "master", iamConfig.URL)
	return t, status, iamErr, err
}

func getToken(username string, password string, clientID string, realm string, iamUrl string) (iamToken, int, iamError, error) {
	var tok iamToken
	var e iamError
	var status = http.StatusInternalServerError

	data := url.Values{}
	data.Set("username", username)
	data.Add("password", password)
	data.Add("client_id", clientID)
	data.Add("grant_type", "password")

	resp, err := http.PostForm(iamUrl+"/realms/"+realm+"/protocol/openid-connect/token", data)
	if err != nil {
		return tok, status, e, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return tok, status, e, err
	}
	if resp.StatusCode != http.StatusOK {
		var e iamError
		json.Unmarshal(body, &e)
		return tok, resp.StatusCode, e, errors.New("failed to get token")
	}
	json.Unmarshal(body, &tok)

	return tok, resp.StatusCode, e, err
}

// registerUser Registers a new user
func registerUser(iamRegReq iamUserRegisterReq, adminToken string, iamConfig config.Iam) (int, iamError, error) {
	var e iamError
	var status = http.StatusInternalServerError
	jsonReq, _ := json.Marshal(iamRegReq)
	req, err := http.NewRequest("POST", iamConfig.URL+"/admin/realms/"+iamConfig.Realm+"/users", bytes.NewBuffer(jsonReq))
	if err != nil {
		return status, e, err
	}

	req.Header.Add("Authorization", "Bearer "+adminToken)
	req.Header.Add(config.ContentTypeHeader, config.ContentTypeJSON)

	//dump, err := httputil.DumpRequest(req, true)
	//dump, err := httputil.DumpRequestOut(req, true)
	//log.Printf("\n %q \n", dump)
	timeout := time.Duration(time.Duration(iamConfig.Timeout) * time.Second)

	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return status, e, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		type errorMsg struct {
			ErrorMessage string `json:"errorMessage"`
		}
		var errMsg errorMsg
		json.Unmarshal(body, &errMsg)
		e.Error = errMsg.ErrorMessage
		e.ErrorType = "Creation failed"
		return resp.StatusCode, e, errors.New("failed to register user")
	}
	return resp.StatusCode, e, err
}

func getUserIamID(userName string, adminToken string, iamConfig config.Iam) (string, int, iamError, error) {
	var userID = ""
	var status = http.StatusInternalServerError
	var e iamError
	req, err := http.NewRequest("GET", iamConfig.URL+"/admin/realms/"+iamConfig.Realm+"/users"+"?username="+userName, nil)
	if err != nil {
		return userID, status, e, err
	}
	//log.Printf("token: %v", t)
	req.Header.Add("Authorization", "Bearer "+adminToken)
	req.Header.Add(config.ContentTypeHeader, config.ContentTypeJSON)

	//dump, err := httputil.DumpRequest(req, true)
	//dump, err := httputil.DumpRequestOut(req, true)
	//log.Printf("\n %q \n", dump)
	timeout := time.Duration(time.Duration(iamConfig.Timeout) * time.Second)

	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return userID, status, e, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		json.Unmarshal(body, &e)
		return userID, resp.StatusCode, e, errors.New("failed to register user")
	}
	type userDetails struct {
		ID string `json:"id"`
	}
	var users []userDetails
	json.Unmarshal(body, &users)
	return users[0].ID, resp.StatusCode, e, err
}

type rReq struct {
	ClientRole         bool   `json:"clientRole"`
	Composite          bool   `json:"composite"`
	ContainerID        string `json:"containerId"`
	Description        string `json:"description"`
	ID                 string `json:"id"`
	Name               string `json:"name"`
	ScopeParamRequired bool   `json:"scopeParamRequired"`
}

var realmRoleOrgAdmin = "70698dc8-3202-4668-a982-4d95107399d4"

func setAdminRole(userID string, adminToken string, iamConfig config.Iam) (int, iamError, error) {
	var status = http.StatusInternalServerError
	var e iamError

	var iReq []rReq
	iReq = append(iReq, rReq{false, false, iamConfig.Realm, "${organization.admin}", realmRoleOrgAdmin, "organization-admin", false})
	jsonReq, _ := json.Marshal(iReq)
	req, err := http.NewRequest("POST", iamConfig.URL+"/admin/realms/"+iamConfig.Realm+"/users/"+userID+"/role-mappings/realm", bytes.NewBuffer(jsonReq))
	if err != nil {
		return status, e, err
	}

	req.Header.Add("Authorization", "Bearer "+adminToken)
	req.Header.Add(config.ContentTypeHeader, config.ContentTypeJSON)

	//dump, err := httputil.DumpRequest(req, true)
	//dump, err := httputil.DumpRequestOut(req, true)
	//fmt.Printf("\n %q \n", dump)
	timeout := time.Duration(time.Duration(iamConfig.Timeout) * time.Second)

	client := http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to set user roles: with status:%v", resp.StatusCode)
		return status, e, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	json.Unmarshal(body, &e)
	return resp.StatusCode, e, err
}

func RegisterUser(regReq config.User, iamConfig config.Iam) (User, error) {
	valid, err := govalidator.ValidateStruct(regReq)
	if !valid {
		log.Printf("Failed to add organization type: Missing mandatory params")
		return User{}, err
	}

	var iamRegReq iamUserRegisterReq
	iamRegReq.Username = regReq.Username
	iamRegReq.Email = regReq.Username
	iamRegReq.Enabled = true

	iamRegReq.Credentials = append(iamRegReq.Credentials, iamCredentials{"password", regReq.Password})
	iamRegReq.RealmRoles = append(iamRegReq.RealmRoles, "organization-admin")

	t, status, iamErr, err := getAdminToken(iamConfig)
	if err != nil {
		log.Printf("Failed to get admin token, user: %v registration", regReq.Username)
		return User{}, err
	}
	fmt.Println(status, iamErr)

	_, _, err = registerUser(iamRegReq, t.AccessToken, iamConfig)
	if err != nil {
		log.Printf("Failed to register user: %v err: %v", regReq.Username, err)
		return User{}, err
	}

	userIamID, _, _, err := getUserIamID(regReq.Username, t.AccessToken, iamConfig)
	if err != nil {
		log.Printf("Failed to get userID for user: %v err: %v", regReq.Username, err)
		return User{}, err
	}

	_, _, err = setAdminRole(userIamID, t.AccessToken, iamConfig)
	if err != nil {
		log.Printf("Failed to set roles for user: %v iam id: %v err: %v", regReq.Username, userIamID, err)
		return User{}, err
	}

	var u User
	u.IamID = userIamID
	u.Email = regReq.Username
	u.Orgs = []Org{}
	u.Roles = []Role{}

	u, err = Add(u)
	if err != nil {
		log.Printf("Failed to add user: %v id: %v to Db err: %v", regReq.Username, userIamID, err)
		return User{}, err
	}

	log.Printf("successfully registered user: %v", regReq.Username)

	return u, err

}

func GetOrganisationAdminToken(userConfig config.User, iamConfig config.Iam) {
	_, _, iamErr, err := getToken(userConfig.Username, userConfig.Password, iamConfig.ClientId, iamConfig.Realm, iamConfig.URL)
	if err != nil {
		log.Printf("Failed to get admin token:%v with error: %s", userConfig.Username, iamErr)
		panic(err)
	}
}
