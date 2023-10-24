package dataattribute

import (
	"context"
	"log"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("dataAttributes")
}

type DataAttribute struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Version        string             `json:"version"`
	AgreementIds   []string           `json:"agreementIds"`
	Name           string             `json:"name" valid:"required"`
	Description    string             `json:"description" valid:"required"`
	Sensitivity    bool               `json:"sensitivity"`
	Category       string             `json:"category"`
	OrganisationId string             `json:"-"`
	IsDeleted      bool               `json:"-"`
}

type DataAgreementForDataAttribute struct {
	Id      string `json:"id" bson:"_id,omitempty"`
	Purpose string `json:"purpose"`
}

type DataAttributeForLists struct {
	Id             primitive.ObjectID              `json:"id" bson:"_id,omitempty"`
	Version        string                          `json:"version"`
	AgreementIds   []string                        `json:"agreementIds"`
	Name           string                          `json:"name" valid:"required"`
	Description    string                          `json:"description" valid:"required"`
	Sensitivity    bool                            `json:"sensitivity"`
	Category       string                          `json:"category"`
	OrganisationId string                          `json:"-"`
	IsDeleted      bool                            `json:"-"`
	AgreementData  []DataAgreementForDataAttribute `json:"dataAgreements"`
}
type DataAttributeRepository struct {
	DefaultFilter bson.M
}

// Init
func (dataAttributeRepo *DataAttributeRepository) Init(organisationId string) {
	dataAttributeRepo.DefaultFilter = bson.M{"organisationid": organisationId, "isdeleted": false}
}

// Add Adds the data attribute to the db
func (dataAttributeRepo *DataAttributeRepository) Add(dataAttribute DataAttribute) (DataAttribute, error) {

	_, err := Collection().InsertOne(context.TODO(), dataAttribute)
	if err != nil {
		return DataAttribute{}, err
	}

	return dataAttribute, nil
}

// Get Gets a single data attribute by given id
func (dataAttributeRepo *DataAttributeRepository) Get(dataAttributeID string) (DataAttribute, error) {
	var result DataAttribute
	dataAttributeId, err := primitive.ObjectIDFromHex(dataAttributeID)
	if err != nil {
		return result, err
	}

	filter := common.CombineFilters(dataAttributeRepo.DefaultFilter, bson.M{"_id": dataAttributeId})

	err = Collection().FindOne(context.TODO(), filter).Decode(&result)

	return result, err
}

// Update Updates the data attribute
func (dataAttributeRepo *DataAttributeRepository) Update(dataAttribute DataAttribute) (DataAttribute, error) {

	filter := common.CombineFilters(dataAttributeRepo.DefaultFilter, bson.M{"_id": dataAttribute.Id})
	update := bson.M{"$set": dataAttribute}

	_, err := Collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return dataAttribute, err
	}
	return dataAttribute, err
}

// Gets data attributes by data agreement id
func (dataAttributeRepo *DataAttributeRepository) GetDataAttributesByDataAgreementId(dataAgreementId string) ([]DataAttribute, error) {
	filter := common.CombineFilters(dataAttributeRepo.DefaultFilter, bson.M{
		"agreementids": bson.M{
			"$in": []string{dataAgreementId},
		},
	})

	cursor, err := Collection().Find(context.Background(), filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())

	var dataAttributes []DataAttribute
	for cursor.Next(context.Background()) {
		var dataAttribute DataAttribute
		if err := cursor.Decode(&dataAttribute); err != nil {
			return dataAttributes, err
		}
		dataAttributes = append(dataAttributes, dataAttribute)
	}

	if err := cursor.Err(); err != nil {
		return dataAttributes, err
	}

	return dataAttributes, err
}

// Removes data agreement id from data attributes
func (dataAttributeRepo *DataAttributeRepository) RemoveDataAgreementIdFromDataAttributes(dataAgreementId string) error {
	filter := common.CombineFilters(dataAttributeRepo.DefaultFilter, bson.M{"agreementids": dataAgreementId})

	update := bson.M{
		"$pull": bson.M{"agreementids": dataAgreementId},
	}

	_, err := Collection().UpdateMany(context.Background(), filter, update)
	if err != nil {
		return err
	}
	err = dataAttributeRepo.DeleteDataAttributesIfDataAgreementIdsIsEmpty()
	if err != nil {
		return err
	}

	return nil
}

// delete data attributes if data agreement ids is empty
func (dataAttributeRepo *DataAttributeRepository) DeleteDataAttributesIfDataAgreementIdsIsEmpty() error {
	filter := common.CombineFilters(dataAttributeRepo.DefaultFilter, bson.M{"agreementids": bson.M{"$exists": true, "$eq": []string{}}})

	// Update to set IsDeleted to true
	update := bson.M{
		"$set": bson.M{
			"isdeleted": true,
		},
	}

	_, err := Collection().UpdateMany(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// ListDataAttributesBasedOnMethodOfUse lists data attributes based on method of use
func ListDataAttributesBasedOnMethodOfUse(methodOfUse string, organisationId string) ([]DataAttributeForLists, error) {
	var results []DataAttributeForLists

	pipeline := []bson.M{
		{"$match": bson.M{"organisationid": organisationId, "isdeleted": false}},
		{"$lookup": bson.M{
			"from": "dataAgreements",
			"let":  bson.M{"localIds": "$agreementids"},
			"pipeline": bson.A{
				bson.M{
					"$match": bson.M{
						"$expr": bson.M{
							"$in": bson.A{"$_id", bson.M{"$map": bson.M{
								"input": "$$localIds",
								"as":    "r",
								"in":    bson.M{"$toObjectId": "$$r"},
							}}},
						},
					},
				},
			},
			"as": "agreementData",
		}},
		{"$match": bson.M{"agreementData.methodofuse": methodOfUse}},
	}

	cursor, err := Collection().Aggregate(context.TODO(), pipeline)
	if err != nil {
		return results, err
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &results); err != nil {
		return results, err
	}
	return results, nil
}

// ListDataAttributesWithDataAgreement lists data attributes with data agreements
func ListDataAttributesWithDataAgreement(organisationId string) ([]DataAttributeForLists, error) {
	var results []DataAttributeForLists

	pipeline := []bson.M{
		{"$match": bson.M{"organisationid": organisationId, "isdeleted": false}},
		{"$lookup": bson.M{
			"from": "dataAgreements",
			"let":  bson.M{"localIds": "$agreementids"},
			"pipeline": bson.A{
				bson.M{
					"$match": bson.M{
						"$expr": bson.M{
							"$in": bson.A{"$_id", bson.M{"$map": bson.M{
								"input": "$$localIds",
								"as":    "r",
								"in":    bson.M{"$toObjectId": "$$r"},
							}}},
						},
					},
				},
			},
			"as": "agreementData",
		}},
	}

	cursor, err := Collection().Aggregate(context.TODO(), pipeline)
	if err != nil {
		return results, err
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &results); err != nil {
		return results, err
	}
	return results, nil
}
