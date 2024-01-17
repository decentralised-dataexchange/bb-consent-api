package jwk

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
)

// JWK represents a JSON Web Key
type JWK struct {
	Kty       string           `json:"kty"`
	Crv       string           `json:"crv"`
	X         string           `json:"x"`
	Y         string           `json:"y"`
	PublicKey *ecdsa.PublicKey `json:"-"`
}

// FromECPublicKey creates JWK from elliptic curve public key
func (obj *JWK) FromECPublicKey(publicKey *ecdsa.PublicKey) *JWK {
	return &JWK{
		Kty:       "EC",
		Crv:       "P-256",
		X:         encodeBase64URL(publicKey.X.Bytes()),
		Y:         encodeBase64URL(publicKey.Y.Bytes()),
		PublicKey: &ecdsa.PublicKey{},
	}
}

// GenerateECKey generate ECDSA key pair using secp256r1 curve
func (obj *JWK) GenerateECKey() *ecdsa.PrivateKey {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Error generating ECDSA key pair: %v", err)
	}

	obj.PublicKey = &privateKey.PublicKey
	return privateKey
}

// ToJSON to json string
func (obj *JWK) ToJSON() string {
	jwk := obj.FromECPublicKey(obj.PublicKey)
	jwkJSON, err := json.MarshalIndent(jwk, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JWK to JSON: %v", err)
	}

	return string(jwkJSON)
}

// ToECPublicKey to ECDSA public key
func (obj *JWK) ToECPublicKey() *ecdsa.PublicKey {
	curve := elliptic.P256() // P-256 curve is assumed, adjust as needed
	xBytes, err := decodeBase64URL(obj.X)
	if err != nil {
		log.Fatalf("failed to decode x parameter: %v", err)
	}

	yBytes, err := decodeBase64URL(obj.Y)
	if err != nil {
		log.Fatalf("failed to decode y parameter: %v", err)
	}

	pubKey := &ecdsa.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}

	return pubKey
}

// FromJSON from json string to JWK
func FromJSON(jwkString string) JWK {
	var jwk JWK
	err := json.Unmarshal([]byte(jwkString), &jwk)
	if err != nil {
		log.Fatal("Failed to unmarshal JWK:", err)
	}
	return jwk
}

// decodeBase64URL use base64.URLEncoding to decode base64 URL-encoded strings
func decodeBase64URL(input string) ([]byte, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(input)
	if err != nil {
		return nil, fmt.Errorf("base64 URL decoding failed: %w", err)
	}
	return decoded, nil
}

// encodeBase64URL encodes the input bytes to base64url format
func encodeBase64URL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
