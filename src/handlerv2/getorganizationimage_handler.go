package handlerv2

import (
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/image"
	"github.com/gorilla/mux"
)

// GetOrganizationImage Retrieves the organization image
func GetOrganizationImage(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Header.Get(config.OrganizationId)
	imageID := mux.Vars(r)["imageID"]

	image, err := image.Get(imageID)

	if err != nil {
		m := fmt.Sprintf("Failed to fetch image with id: %v for org: %v", imageID, organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeImage)
	w.Write(image.Data)
}
