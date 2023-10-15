package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/image"
	"github.com/bb-consent/api/src/org"
)

type coverImageResp struct {
	CoverImageId  string `json:"coverImageId"`
	CoverImageUrl string `json:"coverImageUrl"`
}

// UpdateOrganizationCoverImage Inserts the image and update the id to user
func UpdateOrganizationCoverImage(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Header.Get(config.OrganizationId)

	file, _, err := r.FormFile("orgimage")
	if err != nil {
		m := fmt.Sprintf("Failed to extract image organization: %v", organizationID)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	if err != nil {
		m := fmt.Sprintf("Failed to copy image organization: %v", organizationID)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	imageID, err := image.Add(buf.Bytes())
	if err != nil {
		m := fmt.Sprintf("Failed to store image in data store organization: %v", organizationID)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	imageURL := "https://" + r.Host + "/v1/organizations/" + organizationID + "/image/" + imageID
	o, err := org.UpdateCoverImage(organizationID, imageID, imageURL)
	if err != nil {
		m := fmt.Sprintf("Failed to update organization: %v with image: %v details", organizationID, imageID)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	respBody := coverImageResp{
		CoverImageId:  o.CoverImageID,
		CoverImageUrl: o.CoverImageURL,
	}

	response, _ := json.Marshal(respBody)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
