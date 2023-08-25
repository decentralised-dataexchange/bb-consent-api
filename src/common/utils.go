package common

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
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

// GetRoleID Gets RoleID
func GetRoleID(role string) int {
	for _, r := range orgRoles {
		if r.Role == role {
			return r.ID
		}
	}
	return 0
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
