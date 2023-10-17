package dataagreement

import (
	"context"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/database"
	"github.com/bb-consent/api/src/policy"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("dataAgreements")
}

type Signature struct {
	Id                           string `json:"id"`
	Payload                      string `json:"payload"`
	Signature                    string `json:"signature"`
	VerificationMethod           string `json:"verificationMethod"`
	VerificationPayload          string `json:"verificationPayload"`
	VerificationPayloadHash      string `json:"verificationPayloadHash"`
	VerificationArtifact         string `json:"verificationArtifact"`
	VerificationSignedBy         string `json:"verificationSignedBy"`
	VerificationSignedAs         string `json:"verificationSignedAs"`
	VerificationJwsHeader        string `json:"verificationJwsHeader"`
	Timestamp                    string `json:"timestamp"`
	SignedWithoutObjectReference bool   `json:"signedWithoutObjectReference"`
	ObjectType                   string `json:"objectType"`
	ObjectReference              string `json:"objectReference"`
}

type PolicyForDataAgreement struct {
	policy.Policy
	Id string `json:"id"`
}

type DataAgreement struct {
	Id                      string                 `json:"id" bson:"_id,omitempty"`
	Version                 string                 `json:"version"`
	ControllerId            string                 `json:"controllerId"`
	ControllerUrl           string                 `json:"controllerUrl" valid:"required"`
	ControllerName          string                 `json:"controllerName" valid:"required"`
	Policy                  PolicyForDataAgreement `json:"policy" valid:"required"`
	Purpose                 string                 `json:"purpose" valid:"required"`
	PurposeDescription      string                 `json:"purposeDescription" valid:"required"`
	LawfulBasis             string                 `json:"lawfulBasis" valid:"required"`
	MethodOfUse             string                 `json:"methodOfUse" valid:"required"`
	DpiaDate                string                 `json:"dpiaDate"`
	DpiaSummaryUrl          string                 `json:"dpiaSummaryUrl"`
	Signature               Signature              `json:"signature"`
	Active                  bool                   `json:"active"`
	Forgettable             bool                   `json:"forgettable"`
	CompatibleWithVersionId string                 `json:"compatibleWithVersionId"`
	Lifecycle               string                 `json:"lifecycle" valid:"required"`
	OrganisationId          string                 `json:"-"`
	IsDeleted               bool                   `json:"-"`
}

type DataAgreementRepository struct {
	DefaultFilter bson.M
}

// Init
func (darepo *DataAgreementRepository) Init(organisationId string) {
	darepo.DefaultFilter = bson.M{"organisationid": organisationId, "isdeleted": false}
}

// Add Adds the data agreement to the db
func (darepo *DataAgreementRepository) Add(dataAgreement DataAgreement) (DataAgreement, error) {

	_, err := Collection().InsertOne(context.TODO(), dataAgreement)
	if err != nil {
		return DataAgreement{}, err
	}

	return dataAgreement, nil
}

// Get Gets a single data agreement by given id
func (darepo *DataAgreementRepository) Get(dataAgreementId string) (DataAgreement, error) {

	filter := common.CombineFilters(darepo.DefaultFilter, bson.M{"_id": dataAgreementId})

	var result DataAgreement
	err := Collection().FindOne(context.TODO(), filter).Decode(&result)

	return result, err
}

// Update Updates the data agreement
func (darepo *DataAgreementRepository) Update(dataAgreement DataAgreement) (DataAgreement, error) {

	filter := common.CombineFilters(darepo.DefaultFilter, bson.M{"_id": dataAgreement.Id})
	update := bson.M{"$set": dataAgreement}

	_, err := Collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return dataAgreement, err
	}
	return dataAgreement, err
}
