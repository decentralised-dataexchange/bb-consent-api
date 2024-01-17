package signature

import (
	"time"

	"github.com/bb-consent/api/internal/jwk"
	"github.com/bb-consent/api/internal/jws"
)

type Signature struct {
	Id                           string `json:"id" bson:"_id,omitempty"`
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

// Init
func (s *Signature) Init(ObjectType string, ObjectReference string, SignedWithoutObjectReference bool) {

	s.SignedWithoutObjectReference = SignedWithoutObjectReference
	s.ObjectType = ObjectType
	s.ObjectReference = ObjectReference
	s.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

// CreateSignature
func (s *Signature) CreateSignature(serialisedSnapshot string, serialisedHash string) error {

	s.VerificationPayload = serialisedSnapshot
	s.VerificationPayloadHash = serialisedHash

	return nil

}

// CreateSignatureForConsentRecord
func CreateSignatureForConsentRecord(ObjectType string, ObjectReference string, SignedWithoutObjectReference bool, serialisedSnapshot string, serialisedHash string, signature Signature) (Signature, error) {

	// Create signature
	signature.Init(ObjectType, ObjectReference, SignedWithoutObjectReference)
	err := signature.CreateSignature(serialisedSnapshot, serialisedHash)

	return signature, err
}

// VerifySignature
func VerifySignature(signature string, publicKey string) error {

	jwsObj := jws.JWS{Key: jwk.FromJSON(publicKey), Signature: signature}
	err := jwsObj.Verify()
	if err != nil {
		return err
	}
	return nil
}
