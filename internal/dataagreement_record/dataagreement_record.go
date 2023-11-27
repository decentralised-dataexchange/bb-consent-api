package dataagreementrecord

import (
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DataAgreementRecord struct {
	Id                        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DataAgreementId           string             `json:"dataAgreementId"`
	DataAgreementRevisionId   string             `json:"dataAgreementRevisionId"`
	DataAgreementRevisionHash string             `json:"dataAgreementRevisionHash"`
	IndividualId              string             `json:"individualId"`
	OptIn                     bool               `json:"optIn"`
	State                     string             `json:"state" valid:"required"`
	SignatureId               string             `json:"signatureId"`
	OrganisationId            string             `json:"-"`
	IsDeleted                 bool               `json:"-"`
}

type RevisionForListDataAgreementRecord struct {
	ObjectData string
	Timestamp  string `json:"timestamp"`
}

type DataAgreementRecordForAuditList struct {
	Id        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Revisions []RevisionForListDataAgreementRecord
}

// DataAgreementRecordError is an error enumeration for create consent record API.
type DataAgreementRecordError int

const (
	// IndividualIDIsMissingError indicates that the consent record query params is missing.
	IndividualIdIsMissingError DataAgreementRecordError = iota
	DataAgreementIdIsMissingError
	RevisionIdIsMissingError
	DataAgreementRecordIdIsMissingError
	LawfulBasisIsMissingError
	IdIsMissingError
)

// Error returns the string representation of the error.
func (e DataAgreementRecordError) Error() string {
	switch e {
	case IndividualIdIsMissingError:
		return "Query param individualId is missing!"
	case DataAgreementIdIsMissingError:
		return "Query param  dataAgreementId is missing!"
	case RevisionIdIsMissingError:
		return "Query param revisionId is missing!"
	case DataAgreementRecordIdIsMissingError:
		return "Query param dataAgreementRecordId is missing!"
	case LawfulBasisIsMissingError:
		return "Query param lawfulbasis is missing!"
	case IdIsMissingError:
		return "Query param id is missing!"
	default:
		return "Unknown error!"
	}
}

// ParseQueryParams
func ParseQueryParams(r *http.Request, paramName string, errorType DataAgreementRecordError) (paramValue string, err error) {
	query := r.URL.Query()
	values, ok := query[paramName]
	if ok && len(strings.TrimSpace(values[0])) > 0 {
		return values[0], nil
	}
	return "", errorType
}

type DataAgreementRecordWithTimestamp struct {
	Id                        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DataAgreementId           string             `json:"dataAgreementId"`
	DataAgreementRevisionId   string             `json:"dataAgreementRevisionId"`
	DataAgreementRevisionHash string             `json:"dataAgreementRevisionHash"`
	IndividualId              string             `json:"individualId"`
	OptIn                     bool               `json:"optIn"`
	State                     string             `json:"state" valid:"required"`
	SignatureId               string             `json:"signatureId"`
	Timestamp                 string             `json:"timestamp"`
}
