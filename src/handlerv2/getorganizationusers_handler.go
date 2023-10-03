package handlerv2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/user"
	"github.com/gorilla/mux"
)

type orgUser struct {
	ID    string
	Name  string
	Phone string
	Email string
}
type orgUsers struct {
	Users []orgUser
	Links common.PaginationLinks
}

// GetOrganizationUsers Gets list of organization users
func GetOrganizationUsers(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]

	startID, limit := common.ParsePaginationQueryParameters(r)
	if limit == 0 {
		limit = 20
	}

	users, lastID, err := user.GetOrgSubscribeUsers(organizationID, startID, limit)

	if err != nil {
		m := fmt.Sprintf("Failed to get user subscribed to organization :%v", organizationID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	var ou orgUsers
	for _, u := range users {
		ou.Users = append(ou.Users, orgUser{ID: u.ID.Hex(), Name: u.Name, Phone: u.Phone, Email: u.Email})
	}

	ou.Links = common.CreatePaginationLinks(r, startID, lastID, limit)
	response, _ := json.Marshal(ou)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}
