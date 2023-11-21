package dataagreementrecord

import (
	"context"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/database"
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
func (darRepo *DataAgreementRecordRepository) DeleteAllRecordsForIndividual(individualId string, organisationId string) error {

	filter := bson.M{"organisationid": organisationId, "individualid": individualId}

	_, err := Collection().DeleteMany(context.TODO(), filter)

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

// DataAgreementRecordsWithRevisionsFilteredById This is used for filtering data agreement records by `id` and fetch all revisions
// `id` has 3 possible values - dataAgreementRecordId, dataAgreementId, individualId
func DataAgreementRecordsWithRevisionsFilteredById(organisationId string, id string) ([]DataAgreementRecordForAuditList, error) {

	var pipeline []bson.M

	// Stage 1 - Match by `organisationId` and `isDeleted=false`
	pipeline = append(pipeline, bson.M{"$match": bson.M{"organisationid": organisationId, "isdeleted": false}})

	if len(id) > 0 {

		or := []bson.M{
			{"dataagreementid": id},
			{"individualid": id},
		}

		// Stage 2 - Match `id` against `dataAgreementRecordId`, `dataAgreementId`, `individualId`
		convertIdtoObjectId, err := primitive.ObjectIDFromHex(id)
		if err == nil {
			// Append `dataAgreementRecordId` `or` statements only if
			// string is converted to objectId without errors
			or = append(or, bson.M{"_id": convertIdtoObjectId})
		}

		pipeline = append(pipeline, bson.M{"$match": bson.M{
			"$or": or,
		}})
	}

	// Stage 2 - Lookup revision by `dataAgreementRecordId`
	// This is done to obtain all the revisions of the data agreement record.
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
		},
		"as": "revisions",
	}})
	// Stage 3 - sort by revisions timestamp
	pipeline = append(pipeline, bson.M{"$sort": bson.M{"revisions.timestamp": -1}})

	var results []DataAgreementRecordForAuditList
	cursor, err := Collection().Aggregate(context.Background(), pipeline)
	if err != nil {
		return results, err
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &results); err != nil {
		return results, err
	}

	return results, nil
}

// CreatePipelineForFilteringLatestDataAgreementRecords This pipeline is used for filtering data agreement records
func CreatePipelineForFilteringLatestDataAgreementRecords(organisationId string) ([]primitive.M, error) {

	var pipeline []bson.M

	// Stage 1 - Match by `organisationId` and `isDeleted=false`
	pipeline = append(pipeline, bson.M{"$match": bson.M{"organisationid": organisationId, "isdeleted": false}})

	// Stage 2 - Lookup revision by `dataAgreementRecordId`
	// This is done to obtain timestamp for the latest revision of the data agreement records.
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

// CreatePipelineForFilteringDataAgreementRecordsByIndividualId This pipeline is used for filtering data agreement records by individual id
func CreatePipelineForFilteringDataAgreementRecordsByIndividualId(organisationId string, individualId string) ([]primitive.M, error) {

	var pipeline []bson.M

	// Stage 1 - Match by `organisationId`, `individualId` and `isDeleted=false`
	pipeline = append(pipeline, bson.M{"$match": bson.M{"organisationid": organisationId, "isdeleted": false, "individualid": individualId}})

	// Stage 2 - Lookup revision by `dataAgreementRecordId`
	// This is done to obtain timestamp for the latest revision of the data agreement records.
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

// CreatePipelineForFilteringDataAgreementRecordsByDataAgreementId This pipeline is used for filtering data agreement records by data agreement id
func CreatePipelineForFilteringDataAgreementRecordsByDataAgreementId(organisationId string, dataAgreementId string, individualId string) ([]primitive.M, error) {

	var pipeline []bson.M

	// Stage 1 - Match by `organisationId`, `dataagreementId` and `isDeleted=false`
	pipeline = append(pipeline, bson.M{"$match": bson.M{"organisationid": organisationId, "isdeleted": false, "dataagreementid": dataAgreementId, "individualid": individualId}})

	// Stage 2 - Lookup revision by `dataAgreementRecordId`
	// This is done to obtain timestamp for the latest revision of the data agreement records.
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

// GetAllUsingPipeline
func GetAllUsingPipeline(pipeline []bson.M) ([]DataAgreementRecord, error) {

	var results []DataAgreementRecord
	cursor, err := Collection().Aggregate(context.Background(), pipeline)
	if err != nil {
		return results, err
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &results); err != nil {
		return results, err
	}

	return results, nil
}
