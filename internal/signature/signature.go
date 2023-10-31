package signature

import (
	"encoding/json"

	"github.com/bb-consent/api/internal/common"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Signature struct {
	Id                           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
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

// Init
func (s *Signature) Init(ObjectType string, ObjectReference string, SignedWithoutObjectReference bool) {

	s.SignedWithoutObjectReference = SignedWithoutObjectReference
	s.ObjectType = ObjectType
	s.ObjectReference = ObjectReference
}

// CreateSignature
func (s *Signature) CreateSignature(VerificationPayload interface{}, IsPayload bool) error {

	// Verification Payload
	verificationPayloadSerialised, err := json.Marshal(VerificationPayload)
	if err != nil {
		return err
	}
	s.VerificationPayload = string(verificationPayloadSerialised)

	// Serialised hash using SHA-1
	s.VerificationPayloadHash, err = common.CalculateSHA1(string(verificationPayloadSerialised))
	if err != nil {
		return err
	}

	if IsPayload {
		// Payload
		payload, err := json.Marshal(s)
		if err != nil {
			return err
		}
		s.Payload = string(payload)
	}

	return nil

}

// CreateSignatureForPolicy
func CreateSignatureForObject(ObjectType string, ObjectReference string, SignedWithoutObjectReference bool, VerificationPayload interface{}, IsPayload bool, signature Signature) (Signature, error) {

	// Create signature
	signature.Init(ObjectType, ObjectReference, SignedWithoutObjectReference)
	err := signature.CreateSignature(VerificationPayload, IsPayload)

	return signature, err
}
