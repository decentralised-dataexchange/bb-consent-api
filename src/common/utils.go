package common

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

const (
	// ClientTypeIos IoS client
	ClientTypeIos = 1

	// ClientTypeAndroid Android client
	ClientTypeAndroid = 2

	// ConsentStatusAllow string for consent status
	ConsentStatusAllow = "Allow"

	// ConsentStatusDisAllow string for consent status
	ConsentStatusDisAllow = "Disallow"

	// ConsentStatusAskMe string for consent status
	ConsentStatusAskMe = "AskMe"

	// iGrant Admin role
	iGrantAdminRole = 1000
)

type status struct {
	Code    int
	Message string
}

type statusv2 struct {
	ErrorCode        int    `json:"errorCode"`
	ErrorDescription string `json:"errorDescription"`
}

// OrgRole Organization role definition
type OrgRole struct {
	ID   int
	Role string
}

// Note: Dont change the ID(s) if new role needed then add at the end
var orgRoles = []OrgRole{
	{ID: 1, Role: "Admin"},
	{ID: 2, Role: "Dpo"},
	{ID: 3, Role: "Developer"}}

// PaginationLinks pagination links
type PaginationLinks struct {
	Self string `json:"self,omitempty"`
	Next string `json:"next,omitempty"`
}

// GetRole Gets Role details by ID
func GetRole(roleID int) OrgRole {
	return orgRoles[roleID-1]
}

// GetRoles Gets list of allowed organization roles
func GetRoles() []OrgRole {
	return orgRoles
}

// GetRoleID Gets RoleID
func GetRoleID(role string) int {
	for _, r := range orgRoles {
		if r.Role == role {
			return r.ID
		}
	}
	return 0
}

// IsValidRoleID Check if the role id is valid or not
func IsValidRoleID(roleID int) bool {
	for _, role := range orgRoles {
		if roleID == role.ID {
			return true
		}
	}
	return false
}

// CreatePaginationLinks Creates the self and next links for paginated responses
func CreatePaginationLinks(r *http.Request, startID string, nextID string, limit int) (pagination PaginationLinks) {
	url := "https://" + r.Host + r.URL.Path

	pagination.Self = url + "?limit=" + strconv.Itoa(limit)

	if nextID != "" {
		pagination.Next = url + "?startid=" + nextID + "&limit=" + strconv.Itoa(limit)
	}

	return pagination
}

// ParsePaginationQueryParameters Parses the query parameters that are for pagination
func ParsePaginationQueryParameters(r *http.Request) (startID string, limit int) {
	startID = ""

	startIDs, ok := r.URL.Query()["startid"]

	if ok {
		startID = startIDs[0]
	}

	limits, ok := r.URL.Query()["limit"]

	if ok {
		limit, _ = strconv.Atoi(limits[0])
	} else {
		limit = 10
	}
	return
}

// HandleError Common function to formulate error and set the status
func HandleError(w http.ResponseWriter, code int, message string, err error) {
	s := status{code, message}
	response, _ := json.Marshal(s)

	pc, fn, line, _ := runtime.Caller(1)

	log.Printf("%v with err:%v in %s[%s:%d]", message, err,
		filepath.Base(runtime.FuncForPC(pc).Name()), filepath.Base(fn), line)

	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func HandleErrorV2(w http.ResponseWriter, code int, message string, err error) {
	s := statusv2{
		ErrorCode:        code,
		ErrorDescription: message,
	}
	response, _ := json.Marshal(s)

	pc, fn, line, _ := runtime.Caller(1)

	log.Printf("%v with err:%v in %s[%s:%d]", message, err,
		filepath.Base(runtime.FuncForPC(pc).Name()), filepath.Base(fn), line)

	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// GetRandomString Generate a random alpha numeric string of requested length
func GetRandomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

// Sanitize sanitizes the string
func Sanitize(s string) string {
	p := bluemonday.UGCPolicy()
	return p.Sanitize(s)
}

func IntegerToSemver(version int) string {
	major := version
	minor := 0
	patch := 0

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

func BumpMajorVersion(version string) (string, error) {
	// Split the version into major, minor, and patch
	versionParts := strings.Split(version, ".")

	// Parse the parts into integers
	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return version, err
	}

	// Increment the major version and reset minor and patch to zero
	return fmt.Sprintf("%d.0.0", major+1), err
}

func CalculateSHA1(data string) (string, error) {
	// Convert the JSON string to bytes
	dataBytes := []byte(data)

	// Create a new SHA-1 hasher
	sha1Hasher := sha1.New()

	// Write the data to the hasher
	_, err := sha1Hasher.Write(dataBytes)
	if err != nil {
		return "", err
	}

	// Get the hash sum
	hashSum := sha1Hasher.Sum(nil)

	// Convert the hash sum to a hex string
	hashHex := hex.EncodeToString(hashSum)

	return hashHex, err
}
