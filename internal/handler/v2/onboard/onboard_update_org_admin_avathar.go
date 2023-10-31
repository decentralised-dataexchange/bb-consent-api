package onboard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/image"
	"github.com/bb-consent/api/internal/org"
	"github.com/bb-consent/api/internal/user"
)

type updateOrgAdminImageResp struct {
	Id             string `json:"id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	AvatarImageId  string `json:"avatarImageId"`
	AvatarImageUrl string `json:"avatarImageUrl"`
	LastVisited    string `json:"lastVisited"`
}

// OnboardUpdateOrganisationAdminAvathar
func OnboardUpdateOrganisationAdminAvatar(w http.ResponseWriter, r *http.Request) {
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
		m := "failed to find organisation admin"
		common.HandleErrorV2(w, http.StatusNotFound, m, err)
		return
	}

	file, _, err := r.FormFile("avatarimage")
	if err != nil {
		m := fmt.Sprintf("Failed to extract image user: %v", u.ID.Hex())
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	if err != nil {
		m := fmt.Sprintf("Failed to copy image user: %v", u.ID.Hex())
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	imageID, err := image.Add(buf.Bytes())
	if err != nil {
		m := fmt.Sprintf("Failed to store image in data store user: %v", u.ID.Hex())
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	imageURL := "https://" + r.Host + "/onboard/admin/avatharimage"
	u.ImageID = imageID
	u.ImageURL = imageURL
	u, err = user.Update(u.ID.Hex(), u)
	if err != nil {
		m := fmt.Sprintf("Failed to update user: %v with image: %v details", u.ID, imageID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := updateOrgAdminImageResp{
		Id:             u.ID.Hex(),
		Email:          u.Email,
		Name:           u.Name,
		AvatarImageId:  u.ImageID,
		AvatarImageUrl: u.ImageURL,
		LastVisited:    u.LastVisit,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
