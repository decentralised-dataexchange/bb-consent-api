package onboard

import (
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/image"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/user"
)

// OnboardReadOrganisationAdminAvathar
func OnboardReadOrganisationAdminAvatar(w http.ResponseWriter, r *http.Request) {
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

	image, err := image.Get(u.ImageID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch image with id: %v", u.ImageID)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeImage)
	w.Write(image.Data)
}
