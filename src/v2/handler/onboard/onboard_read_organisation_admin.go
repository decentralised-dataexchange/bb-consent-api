package onboard

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/user"
)

type readOrgAdminResp struct {
	Id             string `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	AvatarImageId  string `json:"avatarImageId"`
	AvatarImageUrl string `json:"avatarImageUrl"`
}

// OnboardReadOrganisationAdmin
func OnboardReadOrganisationAdmin(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	organisation, err := org.Get(organisationId)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization by ID :%v", organisationId)
		common.HandleErrorV2(w, http.StatusNotFound, m, err)
		return
	}

	u, err := user.Get(organisation.Admins[0].UserID)
	if err != nil {
		log.Println("failed to find organisation admin")
		panic(err)
	}
	resp := readOrgAdminResp{
		Id:             u.ID.Hex(),
		Email:          u.Email,
		Name:           u.Name,
		AvatarImageId:  u.ImageID,
		AvatarImageUrl: u.ImageURL,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
