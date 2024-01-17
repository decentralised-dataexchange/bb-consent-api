package jws

import (
	"fmt"

	"github.com/bb-consent/api/internal/jwk"
	"github.com/go-jose/go-jose/v3"
)

// JWS represents a JSON Web Signature (JWS)
type JWS struct {
	Claims    string  `json:"-"`
	Key       jwk.JWK `json:"-"`
	Signature string  `json:"-"`
}

// Verify verify JSON web signature (JWS)
func (obj *JWS) Verify() error {
	// Create EC public key
	pubKey := obj.Key.ToECPublicKey()
	// Deserialise to JWS
	jws, err := jose.ParseSigned(obj.Signature)
	if err != nil {
		return err
	}
	// Verify signature and return decoded payload
	payload, err := jws.Verify(pubKey)
	if err != nil {
		return err
	}
	// Print decoded payload
	fmt.Println("Signature verified.")
	fmt.Printf("\nPayload: \n\n%v\n", string(payload))

	return nil
}
