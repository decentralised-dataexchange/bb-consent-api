package apikey

import (
	"errors"
	"fmt"
	"time"

	"github.com/bb-consent/api/internal/config"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ApiKey struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name           string             `json:"name"`
	Scopes         []string           `json:"scopes" valid:"required"`
	Apikey         string             `json:"apiKey"`
	ExpiryInDays   int                `json:"expiryInDays"`
	OrganisationId string             `json:"-"`
	IsDeleted      bool               `json:"-"`
	Timestamp      string             `json:"-"`
}

var ApiSecretKey string

func Init(config *config.Configuration) {
	ApiSecretKey = config.ApiSecretKey

}

type Claims struct {
	Scopes              []string
	OrganisationId      string
	OrganisationAdminId string
	// Add other fields as needed
	jwt.StandardClaims
}

// Create Create apikey
func Create(scopes []string, expiresAt int64, organisationId string, organisationAdminId string) (string, error) {
	var SigningKey = []byte(ApiSecretKey)
	// Create the Claims
	claims := Claims{
		scopes,
		organisationId,
		organisationAdminId,
		jwt.StandardClaims{

			ExpiresAt: expiresAt,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(SigningKey)

	return ss, err
}

// Decode Decodes the apikey
func Decode(apiKey string) (claims Claims, err error) {
	var SigningKey = []byte(ApiSecretKey)
	token, err := jwt.ParseWithClaims(apiKey, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return SigningKey, nil
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
