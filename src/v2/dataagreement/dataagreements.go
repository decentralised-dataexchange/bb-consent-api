package dataagreement

import (
	"context"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/database"
	"github.com/bb-consent/api/src/policy"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("dataAgreements")
}

type DataAttribute struct {
	Id          primitive.ObjectID `json:"id" bson:"id,omitempty"`
	Name        string             `json:"name" valid:"required"`
	Description string             `json:"description" valid:"required"`
	Sensitivity bool               `json:"sensitivity"`
	Category    string             `json:"category"`
}

type Signature struct {
	Id                           primitive.ObjectID `json:"id"`
	Payload                      string             `json:"payload"`
	Signature                    string             `json:"signature"`
	VerificationMethod           string             `json:"verificationMethod"`
	VerificationPayload          string             `json:"verificationPayload"`
	VerificationPayloadHash      string             `json:"verificationPayloadHash"`
	VerificationArtifact         string             `json:"verificationArtifact"`
	VerificationSignedBy         string             `json:"verificationSignedBy"`
	VerificationSignedAs         string             `json:"verificationSignedAs"`
	VerificationJwsHeader        string             `json:"verificationJwsHeader"`
	Timestamp                    string             `json:"timestamp"`
	SignedWithoutObjectReference bool               `json:"signedWithoutObjectReference"`
	ObjectType                   string             `json:"objectType"`
	ObjectReference              string             `json:"objectReference"`
}

type PolicyForDataAgreement struct {
	policy.Policy
	Id string `json:"id"`
}

type DataAgreement struct {
	Id                      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Version                 string             `json:"version"`
	ControllerId            string             `json:"controllerId"`
	ControllerUrl           string             `json:"controllerUrl" valid:"required"`
	ControllerName          string             `json:"controllerName" valid:"required"`
	Policy                  policy.Policy      `json:"policy" valid:"required"`
	Purpose                 string             `json:"purpose" valid:"required"`
	PurposeDescription      string             `json:"purposeDescription" valid:"required"`
	LawfulBasis             string             `json:"lawfulBasis" valid:"required"`
	MethodOfUse             string             `json:"methodOfUse" valid:"required"`
	DpiaDate                string             `json:"dpiaDate"`
	DpiaSummaryUrl          string             `json:"dpiaSummaryUrl"`
	Signature               Signature          `json:"signature"`
	Active                  bool               `json:"active"`
	Forgettable             bool               `json:"forgettable"`
	CompatibleWithVersionId string             `json:"compatibleWithVersionId"`
	Lifecycle               string             `json:"lifecycle" valid:"required"`
	DataAttributes          []DataAttribute    `json:"dataAttributes" valid:"required"`
	OrganisationId          string             `json:"-"`
	IsDeleted               bool               `json:"-"`
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
func (darepo *DataAgreementRepository) Get(dataAgreementID string) (DataAgreement, error) {
	dataAgreementId, err := primitive.ObjectIDFromHex(dataAgreementID)
	if err != nil {
		return DataAgreement{}, err
	}

	filter := common.CombineFilters(darepo.DefaultFilter, bson.M{"_id": dataAgreementId})

	var result DataAgreement
	err = Collection().FindOne(context.TODO(), filter).Decode(&result)
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
	return dataAgreement, nil
}

// IsDataAgreementExist Check if data agreement with given id exists
func (darepo *DataAgreementRepository) IsDataAgreementExist(dataAgreementID string) (int64, error) {
	var exists int64
	dataAgreementId, err := primitive.ObjectIDFromHex(dataAgreementID)
	if err != nil {
		return exists, err
	}

	filter := common.CombineFilters(darepo.DefaultFilter, bson.M{"_id": dataAgreementId})

	exists, err = Collection().CountDocuments(context.TODO(), filter)
	if err != nil {
		return exists, err
	}
	return exists, nil
}

// CreatePipelineForFilteringDataAgreements This pipeline is used for filtering data agreements
func CreatePipelineForFilteringDataAgreements(organisationId string) ([]primitive.M, error) {

	var pipeline []bson.M

	// Stage 1 - Match by `organisationId` and `isDeleted=false`
	pipeline = append(pipeline, bson.M{"$match": bson.M{"organisationid": organisationId, "isdeleted": false}})

	// Stage 2 - Lookup revision by `dataAgreementId`
	// This is done to obtain timestamp for the latest revision of the data agreements.
	pipeline = append(pipeline, bson.M{"$lookup": bson.M{
		"from": "revisions",
		"let":  bson.M{"localId": "$_id"},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$eq": []interface{}{"$objectid", bson.M{"$toString": "$$localId"}},
					},
				},
			},
			bson.M{
				"$sort": bson.M{"timestamp": -1},
			},
			bson.M{"$limit": int64(1)},
		},
		"as": "revisions",
	}})

	// Stage 3 - Add the timestamp from revisions
	pipeline = append(pipeline, bson.M{"$addFields": bson.M{"timestamp": bson.M{
		"$let": bson.M{
			"vars": bson.M{
				"first": bson.M{
					"$arrayElemAt": bson.A{"$revisions", 0},
				},
			},
			"in": "$$first.timestamp",
		},
	}}})

	// Stage 4 - Remove revisions field
	pipeline = append(pipeline, bson.M{
		"$project": bson.M{
			"revisions": 0,
		},
	})

	return pipeline, nil
}

// CreatePipelineForFilteringDataAgreements This pipeline is used for filtering data agreements
func CreatePipelineForFilteringDataAgreementsUsingLifecycle(organisationId string, lifecycle string) ([]primitive.M, error) {

	var pipeline []bson.M

	// Stage 1 - Match by `organisationId` and `isDeleted=false` and lifecycle
	pipeline = append(pipeline, bson.M{"$match": bson.M{"organisationid": organisationId, "isdeleted": false, "lifecycle": lifecycle}})

	// Stage 2 - Lookup revision by `dataAgreementId`
	// This is done to obtain timestamp for the latest revision of the data agreements.
	pipeline = append(pipeline, bson.M{"$lookup": bson.M{
		"from": "revisions",
		"let":  bson.M{"localId": "$_id"},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$eq": []interface{}{"$objectid", bson.M{"$toString": "$$localId"}},
					},
				},
			},
			bson.M{
				"$sort": bson.M{"timestamp": -1},
			},
			bson.M{"$limit": int64(1)},
		},
		"as": "revisions",
	}})

	// Stage 3 - Add the timestamp from revisions
	pipeline = append(pipeline, bson.M{"$addFields": bson.M{"timestamp": bson.M{
		"$let": bson.M{
			"vars": bson.M{
				"first": bson.M{
					"$arrayElemAt": bson.A{"$revisions", 0},
				},
			},
			"in": "$$first.timestamp",
		},
	}}})

	// Stage 4 - Remove revisions field
	pipeline = append(pipeline, bson.M{
		"$project": bson.M{
			"revisions": 0,
		},
	})

	return pipeline, nil
}

// GetDataAttributeById Gets a single data agreement by data attribute id
func (darepo *DataAgreementRepository) GetByDataAttributeId(dataAttributeID string) (DataAgreement, error) {
	dataAgreementId, err := primitive.ObjectIDFromHex(dataAttributeID)
	if err != nil {
		return DataAgreement{}, err
	}

	filter := common.CombineFilters(darepo.DefaultFilter, bson.M{"dataattributes.id": dataAgreementId})

	var result DataAgreement
	err = Collection().FindOne(context.TODO(), filter).Decode(&result)
	return result, err
}

// GetByMethodOfUse Gets data agreements by method of use
func (darepo *DataAgreementRepository) GetByMethodOfUse(methodOfUse string) ([]DataAgreement, error) {

	filter := common.CombineFilters(darepo.DefaultFilter, bson.M{"methodofuse": methodOfUse})

	var results []DataAgreement
	cursor, err := Collection().Find(context.TODO(), filter)
	if err != nil {
		return results, err
	}

	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		return results, err
	}
	return results, nil
}

// GetAll Gets all data agreements
func (darepo *DataAgreementRepository) GetAll() ([]DataAgreement, error) {

	var results []DataAgreement
	cursor, err := Collection().Find(context.TODO(), darepo.DefaultFilter)
	if err != nil {
		return results, err
	}

	defer cursor.Close(context.TODO())

	if err := cursor.All(context.TODO(), &results); err != nil {
		return results, err
	}
	return results, nil
}
