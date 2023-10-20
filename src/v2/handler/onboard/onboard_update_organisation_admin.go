package onboard

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/user"
	"github.com/bb-consent/api/src/v2/iam"
)

type updateOrgAdminReq struct {
	Id             string `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name" valid:"required"`
	AvatarImageId  string `json:"avatarImageId"`
	AvatarImageUrl string `json:"avatarImageUrl"`
}

type updateOrgAdminResp struct {
	Id             string `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	AvatarImageId  string `json:"avatarImageId"`
	AvatarImageUrl string `json:"avatarImageUrl"`
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

	if strings.TrimSpace(upReq.Name) != "" {
		u.Name = upReq.Name

		err = iam.UpdateIamUser(upReq.Name, u.IamID)
		if err != nil {
			m := fmt.Sprintf("Failed to update IAM user by id:%v", u.ID)
			common.HandleError(w, http.StatusInternalServerError, m, err)
			return
		}

	}

	u, err = user.Update(u.ID.Hex(), u)
	if err != nil {
		m := fmt.Sprintf("Failed to update user by id:%v", u.ID.Hex())
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := updateOrgAdminResp{
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
