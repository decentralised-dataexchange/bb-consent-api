package service

import (
	"fmt"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/image"
	"github.com/gorilla/mux"
)

// ServiceReadOrganisationImage Retrieves the organization image
func ServiceReadOrganisationImage(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Header.Get(config.OrganizationId)
	// Path params
	imageId := mux.Vars(r)["imageId"]

	image, err := image.Get(imageId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch image with id: %v for org: %v", imageId, organizationID)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeImage)
	w.Write(image.Data)
}
