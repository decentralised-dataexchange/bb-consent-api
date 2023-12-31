package service

import (
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/image"
	"github.com/bb-consent/api/internal/org"
)

// ServiceReadOrganisationCoverImage Retrieves the organization cover image
func ServiceReadOrganisationCoverImage(w http.ResponseWriter, r *http.Request) {
	organizationID := common.Sanitize(r.Header.Get(config.OrganizationId))
	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization by ID :%v", organizationID)
		common.HandleErrorV2(w, http.StatusNotFound, m, err)
		return
	}
	imageID := o.CoverImageID

	image, err := image.Get(imageID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch image with id: %v for org: %v", imageID, organizationID)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeImage)
	w.Write(image.Data)
}
