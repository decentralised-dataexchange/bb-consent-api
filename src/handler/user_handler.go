package handler

import (
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
