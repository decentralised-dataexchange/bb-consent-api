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

// CountDataAgreementRecords counts the data agreement record containing data agreement id and individual id
func (darRepo *DataAgreementRecordRepository) CountDataAgreementRecords(dataAgreementId string, individualId string) (int64, error) {
	filter := common.CombineFilters(darRepo.DefaultFilter, bson.M{"individualid": individualId, "dataagreementid": dataAgreementId})

	count, err := Collection().CountDocuments(context.Background(), filter)
	if err != nil {
		return count, nil
	}

	return count, nil
}

// PipelineForList creates pipeline for list data agreement records
func PipelineForList(organisationId string, id string, lawfulBasis string, isId bool, isLawfulBasis bool) ([]primitive.M, error) {
	var pipeline []primitive.M

	var pipelineForIdExists []primitive.M
	if isId {
		dataAgreementRecordId, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return []bson.M{}, err
		}

		pipelineForIdExists = []bson.M{
			{"$match": bson.M{
				"$or": []bson.M{
					{"_id": dataAgreementRecordId},
					{"dataagreementid": id},
					{"individualid": id},
				},
			},
			},
		}
	}
	lookupAgreementStage := bson.M{"$lookup": bson.M{
		"from": "dataAgreements",
		"let":  bson.M{"localId": "$dataagreementid"},
		"pipeline": bson.A{
			bson.M{
				"$match": bson.M{
					"$expr": bson.M{
						"$eq": []interface{}{"$_id", bson.M{"$toObjectId": "$$localId"}},
					},
				},
			},
		},
		"as": "dataAgreements",
	}}
	unwindStage := bson.M{"$unwind": "$dataAgreements"}
	lookupRevisionStage := bson.M{"$lookup": bson.M{
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
		},
		"as": "Revisions",
	}}
	addRevisionFieldStage := bson.M{
		"$addFields": bson.M{
			"Revision": bson.M{
				"$arrayElemAt": []interface{}{
					bson.M{
						"$slice": []interface{}{"$Revisions", -1},
					},
					0,
				},
			},
		},
	}
	addTimestampFieldStage := bson.M{"$addFields": bson.M{"timestamp": "$Revision.timestamp"}}
	projectStage := bson.M{
		"$project": bson.M{
			"Revisions": 0,
			"Revision":  0,
		},
	}

	pipelineForIdNotExists := []bson.M{
		{"$match": bson.M{"organisationid": organisationId, "isdeleted": false}},
	}

	lawfulBasisMatch := bson.M{
		"$match": bson.M{
			"dataAgreements.lawfulbasis": lawfulBasis,
		},
	}
	if isId && isLawfulBasis {
		pipeline = append(pipelineForIdExists, lookupAgreementStage, unwindStage, lookupRevisionStage, addRevisionFieldStage, addTimestampFieldStage, projectStage, lawfulBasisMatch)
	} else if isId && !isLawfulBasis {
		pipeline = append(pipelineForIdExists, lookupAgreementStage, unwindStage, lookupRevisionStage, addRevisionFieldStage, addTimestampFieldStage, projectStage)

	} else if isLawfulBasis && !isId {
		pipeline = append(pipelineForIdNotExists, lookupAgreementStage, unwindStage, lookupRevisionStage, addRevisionFieldStage, addTimestampFieldStage, projectStage, lawfulBasisMatch)
	} else {
		pipeline = []bson.M{}
	}

	return pipeline, nil
}
