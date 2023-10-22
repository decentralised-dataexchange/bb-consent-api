package dataagreementrecord

import (
	"context"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("dataAgreementRecords")
}

type DataAgreementRecordRepository struct {
	DefaultFilter bson.M
}

// Init
func (darRepo *DataAgreementRecordRepository) Init(organisationId string) {
	darRepo.DefaultFilter = bson.M{"organisationid": organisationId, "isdeleted": false}
}

// Add Adds the data agreement record to the db
func (darRepo *DataAgreementRecordRepository) Add(dataAgreementRecord DataAgreementRecord) (DataAgreementRecord, error) {

	_, err := Collection().InsertOne(context.TODO(), dataAgreementRecord)
	if err != nil {
		return DataAgreementRecord{}, err
	}

	return dataAgreementRecord, nil
}

// Get Gets a single data agreement record
func (darRepo *DataAgreementRecordRepository) Get(dataAgreementRecordID string) (DataAgreementRecord, error) {
	dataAgreementRecordId, err := primitive.ObjectIDFromHex(dataAgreementRecordID)
	if err != nil {
		return DataAgreementRecord{}, err
	}

	filter := common.CombineFilters(darRepo.DefaultFilter, bson.M{"_id": dataAgreementRecordId})

	var result DataAgreementRecord
	err = Collection().FindOne(context.TODO(), filter).Decode(&result)

	return result, err
}

// Update Updates the data agreement record
func (darRepo *DataAgreementRecordRepository) Update(dataAgreementRecord DataAgreementRecord) (DataAgreementRecord, error) {

	filter := common.CombineFilters(darRepo.DefaultFilter, bson.M{"_id": dataAgreementRecord.Id})
	update := bson.M{"$set": dataAgreementRecord}

	_, err := Collection().UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return dataAgreementRecord, err
	}
	return dataAgreementRecord, err
}

// Get Gets a single data agreement record by data agreement id and individual id
func (darRepo *DataAgreementRecordRepository) GetByDataAgreementIdandIndividualId(dataAgreementId string, individualId string) (DataAgreementRecord, error) {

	filter := common.CombineFilters(darRepo.DefaultFilter, bson.M{"individualid": individualId, "dataagreementid": dataAgreementId})

	var result DataAgreementRecord
	err := Collection().FindOne(context.TODO(), filter).Decode(&result)

	return result, err
}

// Deletes all the data agreement records of individual
func (darRepo *DataAgreementRecordRepository) DeleteAllRecordsForIndividual(individualId string) error {

	filter := common.CombineFilters(darRepo.DefaultFilter, bson.M{"individualid": individualId})

	// Update to set IsDeleted to true
	update := bson.M{
		"$set": bson.M{
			"isdeleted": true,
		},
	}

	_, err := Collection().UpdateMany(context.TODO(), filter, update)

	return err
}
