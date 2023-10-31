package dataagreementrecordhistory

import (
	"context"
	"time"

	"github.com/bb-consent/api/internal/database"
	"go.mongodb.org/mongo-driver/mongo"
)

func Collection() *mongo.Collection {
	return database.DB.Client.Database(database.DB.Name).Collection("dataAgreementRecordsHistories")
}

// Add Adds the Data Agreement Records History to the db
func Add(dataAgreementRecordsHistory DataAgreementRecordsHistory) (DataAgreementRecordsHistory, error) {

	dataAgreementRecordsHistory.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	_, err := Collection().InsertOne(context.TODO(), dataAgreementRecordsHistory)
	if err != nil {
		return DataAgreementRecordsHistory{}, err
	}

	return dataAgreementRecordsHistory, nil
}
