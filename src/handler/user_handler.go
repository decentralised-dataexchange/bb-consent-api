package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/user"
)

// GetUserByIamID Gets a single user by given Iamid
func GetUserByIamID(iamID string) (user.User, error) {
	user, err := user.GetByIamID(iamID)

	if err != nil {
		return user, err
	}
	return user, err
}

// GetUserRoles Get User roles as int array
func GetUserRoles(userRoles []user.Role) (roles []int) {
	for _, role := range userRoles {
		roles = append(roles, role.RoleID)
	}
	return
}

// GetUser Gets a single user by ID
func GetUser(userID string) (user.User, error) {
	user, err := user.Get(userID)

	if err != nil {
		return user, err
	}
	return user, err
}

type userResp struct {
	User user.User
}

// GetCurrentUser Gets the currernt authenticated user details
func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := token.GetUserID(r)

	u, err := user.Get(userID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch user by id:%v", userID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(userResp{u})
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
