package onboard

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/iam"
	"github.com/bb-consent/api/internal/org"
	"github.com/bb-consent/api/internal/user"
)

type orgAdmin struct {
	Id             string `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name" valid:"required"`
	AvatarImageId  string `json:"avatarImageId"`
	AvatarImageUrl string `json:"avatarImageUrl"`
	LastVisited    string `json:"lastVisited"`
}

type updateOrgAdminReq struct {
	OrganisationAdmin orgAdmin `json:"organisationAdmin"`
}

type updateOrgAdminResp struct {
	OrganisationAdmin orgAdmin `json:"organisationAdmin"`
}

// OnboardUpdateOrganisationAdmin
func OnboardUpdateOrganisationAdmin(w http.ResponseWriter, r *http.Request) {
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

	var upReq updateOrgAdminReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &upReq)

	// validating request payload
	valid, err := govalidator.ValidateStruct(upReq)
	if !valid {
		log.Printf("Missing mandatory params for updating organization admin")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	if strings.TrimSpace(upReq.OrganisationAdmin.Name) != "" {
		u.Name = upReq.OrganisationAdmin.Name

		err = iam.UpdateIamUser(upReq.OrganisationAdmin.Name, u.IamID)
		if err != nil {
			m := fmt.Sprintf("Failed to update IAM user by id:%v", u.ID)
			common.HandleError(w, http.StatusInternalServerError, m, err)
			return
		}

	}

	u, err = user.Update(u.ID, u)
	if err != nil {
		m := fmt.Sprintf("Failed to update user by id:%v", u.ID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	orgAdmin := orgAdmin{
		Id:             u.ID,
		Email:          u.Email,
		Name:           u.Name,
		AvatarImageId:  u.ImageID,
		AvatarImageUrl: u.ImageURL,
		LastVisited:    u.LastVisit,
	}

	response, _ := json.Marshal(updateOrgAdminResp{OrganisationAdmin: orgAdmin})
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
