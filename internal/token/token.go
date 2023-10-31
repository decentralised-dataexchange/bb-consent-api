package token

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/bb-consent/api/internal/config"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
)

const (
	// AuthorizationUnknown Unknwon type
	AuthorizationUnknown = 1
	// AuthorizationAPIKey Uses apikey for API access
	AuthorizationAPIKey = 2
	// AuthorizationToken Uses jwt token for API access
	AuthorizationToken = 3
)

// JWKS.json from iam.igrant.io
var jwks config.JSONWebKeys

// Init Initialize the IAM handler
func Init(config *config.Configuration) {
	jwks = config.Iam.Jwks
}

// DecodeAuthHeader Decodes Authorization header and returns type and key
func DecodeAuthHeader(r *http.Request) (authType int, key string, err error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return AuthorizationUnknown, "", errors.New("Missing Auth header")
	}
	// Check for Bearer prefix in the header
	primaryToken := strings.TrimPrefix(authHeader, "Bearer ")
	if primaryToken != authHeader && len(primaryToken) > 0 {
		return AuthorizationToken, primaryToken, nil
	}

	// Check for ApiKey prefix in the header
	apiKey := strings.TrimPrefix(authHeader, "ApiKey ")
	if apiKey != authHeader && len(apiKey) > 0 {
		return AuthorizationAPIKey, apiKey, nil
	}

	return AuthorizationUnknown, "", errors.New("Incorrect Auth header format")
}

type roles struct {
	Roles []string
}

// AccessToken the token struct filled after decoding the token
type AccessToken struct {
	IamID             string `json:"sub"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	Exp               uint32 `json:"exp"`
	Iat               uint32 `json:"iat"`
	Kid               string `json:"kid"`
	Alg               string `json:"alg"`
	RealmAccess       roles  `json:"realm_access"`
	PreferredUsername string `json:"preferred_username"`
	jwt.StandardClaims
}

// ParseToken parses the token and returns the accessToken struct
func ParseToken(tokenString string) (AccessToken, error) {
	accToken := AccessToken{}
	decodedE, err := base64.RawURLEncoding.DecodeString(jwks.RsaRawE)
	if err != nil {
		return accToken, err
	}
	if len(decodedE) < 4 {
		ndata := make([]byte, 4)
		copy(ndata[4-len(decodedE):], decodedE)
		decodedE = ndata
	}
	pubKey := &rsa.PublicKey{
		N: &big.Int{},
		E: int(binary.BigEndian.Uint32(decodedE[:])),
	}
	decodedN, err := base64.RawURLEncoding.DecodeString(jwks.RsaRawN)
	if err != nil {
		return accToken, err
	}
	pubKey.N.SetBytes(decodedN)

	token, err := jwt.ParseWithClaims(tokenString, &accToken, func(token *jwt.Token) (verifykey interface{}, err error) {
		return pubKey, nil
	})

	if err != nil || token.Valid != true {
		return accToken, err
	}

	//Check the token expiry
	if int64(accToken.Exp) < time.Now().Unix() {
		return accToken, errors.New("Token expired")
	}
	return accToken, nil
}

const tokenKey = "token"
const userIDKey = "userID"
const rolesKey = "roles"
const APIKey = "apiKey"
const UserRoleKey = "role"

// Set Set the token to context
func Set(r *http.Request, token AccessToken) {
	context.Set(r, tokenKey, token)
}

// Get Get the token from context
func Get(r *http.Request) AccessToken {
	return context.Get(r, tokenKey).(AccessToken)
}

// GetIamID Get the iam id from context
func GetIamID(r *http.Request) string {
	t := context.Get(r, tokenKey).(AccessToken)
	return t.IamID
}

// GetUserName Get UserName from context
func GetUserName(r *http.Request) string {
	t := context.Get(r, tokenKey).(AccessToken)
	return t.Email
}

// SetUserID Set UserID to context
func SetUserID(r *http.Request, userID string) {
	context.Set(r, userIDKey, userID)
}

// GetUserID Get UserID from context
func GetUserID(r *http.Request) string {
	if context.Get(r, userIDKey) == nil {
		return ""
	}
	return context.Get(r, userIDKey).(string)
}

// SetUserRoles Set Roles to context
func SetUserRoles(r *http.Request, roles []int) {
	context.Set(r, rolesKey, roles)
}

// GetUserRoles Get Roles from context
func GetUserRoles(r *http.Request) []int {
	return context.Get(r, rolesKey).([]int)
}

// SetUserAPIKey Set API Key to context
func SetUserAPIKey(r *http.Request, apiKey string) {
	context.Set(r, APIKey, apiKey)
}

// GetUserAPIKey Get API key from context
func GetUserAPIKey(r *http.Request) string {
	return context.Get(r, APIKey).(string)
}

// IsOrgAdmin Is organization admin
func IsOrgAdmin(r *http.Request) bool {
	t := context.Get(r, tokenKey).(AccessToken)
	for i := range t.RealmAccess.Roles {
		if t.RealmAccess.Roles[i] == "organization-admin" {
			return true
		}
	}
	return false
}

func SetUserRole(r *http.Request, userRole string) {
	context.Set(r, UserRoleKey, userRole)
}

func GetUserRole(r *http.Request) string {
	return context.Get(r, UserRoleKey).(string)
}

func SetUserToRequestContext(r *http.Request, userID string, userRole string) {
	context.Set(r, userIDKey, userID)
	context.Set(r, UserRoleKey, userRole)

	// Set individual to request header if not present
	if _, exists := r.Header[http.CanonicalHeaderKey(config.IndividualHeaderKey)]; !exists {
		r.Header.Set(config.IndividualHeaderKey, userID)
	}

}
