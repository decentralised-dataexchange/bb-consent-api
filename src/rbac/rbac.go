package rbac

import (
	"github.com/bb-consent/api/src/user"
)

// RBAC User Roles
const (
	ROLE_USER  string = "user"
	ROLE_ADMIN string = "organisation_admin"
)

// IsOrgAdmin   is user an admin in the organisation
func IsOrgAdmin(roles []user.Role, orgID string) bool {
	for _, item := range roles {
		if item.RoleID == 1 {
			if item.OrgID == orgID {
				return true
			}
		}
	}
	return false
}

// IsUser  is User Role user
func IsUser(roles []user.Role) bool {
	return len(roles) == 0
}
