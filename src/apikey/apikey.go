package apikey

import (
	"errors"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// Claims apikey claims
type Claims struct {
	UserID string `json:"userid"`
	OrgID  string `json:"orgid"`
	Env    string `json:"env"`
	jwt.StandardClaims
}

var mySigningKey = []byte("sample")

// Decode Decodes the apikey
func Decode(apiKey string) (claims Claims, err error) {
	token, err := jwt.ParseWithClaims(apiKey, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return mySigningKey, nil
	})

	if err != nil || !token.Valid {
		return claims, err
	}

	//Check the token expiry
	if int64(claims.StandardClaims.ExpiresAt) < time.Now().Unix() {
		return claims, errors.New("token expired")
	}
	return claims, nil
}
