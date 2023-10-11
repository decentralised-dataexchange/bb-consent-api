package handlerv2

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type organizationResp struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string             `json:"name"`
	Location      string             `json:"location"`
	PolicyURL     string             `json:"policyUrl"`
	CoverImageID  string             `json:"coverImageId"`
	CoverImageURL string             `json:"coverImageUrl"`
	LogoImageID   string             `json:"logoImageId"`
	LogoImageURL  string             `json:"logoImageUrl"`
}

type getOrgResp struct {
	Organization organizationResp `json:"organization"`
}

// GetOrganizationByID Gets a single organization by given id
func GetOrganizationByID(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Header.Get(config.OrganizationId)
	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization by ID :%v", organizationID)
		common.HandleErrorV2(w, http.StatusNotFound, m, err)
		return
	}

	oResp := organizationResp{
		ID:            o.ID,
		Name:          o.Name,
		Location:      o.Location,
		PolicyURL:     o.PolicyURL,
		CoverImageID:  o.CoverImageID,
		CoverImageURL: o.CoverImageURL,
		LogoImageID:   o.LogoImageID,
		LogoImageURL:  o.LogoImageURL,
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	response, _ := json.Marshal(getOrgResp{oResp})
	w.Write(response)
}
