package dataagreementrecordhistory

import (
	"fmt"
	"log"

	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/org"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DataAgreementRecordsHistory
type DataAgreementRecordsHistory struct {
	Id              string `json:"id" bson:"_id,omitempty"`
	OrganisationId  string `json:"organisationId"`
	DataAgreementId string `json:"dataAgreementId"`
	Log             string `json:"log"`
	Timestamp       string `json:"timestamp"`
	ConsentRecordId string `json:"consentRecordId"`
	IndividualId    string `json:"individualId"`
}

func DataAgreementRecordHistoryAdd(darH DataAgreementRecordsHistory, optIn bool) error {
	o, err := org.Get(darH.OrganisationId)
	if err != nil {
		return err
	}
	// Repository
	darepo := dataagreement.DataAgreementRepository{}
	darepo.Init(darH.OrganisationId)

	dataAgreement, err := darepo.Get(darH.DataAgreementId)
	if err != nil {
		return err
	}

	if optIn {
		value := "Allow"
		darH.Log = fmt.Sprintf("Updated consent value to <%s> for the purpose <%s> in organization <%s>",
			value, dataAgreement.Purpose, o.Name)
	} else {
		value := "Disallow"
		darH.Log = fmt.Sprintf("Updated consent value to <%s> for the purpose <%s> in organization <%s>",
			value, dataAgreement.Purpose, o.Name)
	}
	log.Printf("The log is: %s", darH.Log)

	darH.Id = primitive.NewObjectID().Hex()

	_, err = Add(darH)
	if err != nil {
		return err
	}
	return nil

}
