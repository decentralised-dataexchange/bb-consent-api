package onboard

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/image"
	"github.com/bb-consent/api/internal/org"
)

// UpdateOrganizationLogoImage Inserts the image and update the id to user
func UpdateOrganizationLogoImage(w http.ResponseWriter, r *http.Request) {
	organizationID := common.Sanitize(r.Header.Get(config.OrganizationId))

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

	_, err = org.UpdateLogoImage(organizationID, imageID)
	if err != nil {
		m := fmt.Sprintf("Failed to update organization: %v with image: %v details", organizationID, imageID)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
}
