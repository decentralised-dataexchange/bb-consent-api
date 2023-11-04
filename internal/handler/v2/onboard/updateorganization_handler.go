package onboard

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/org"
	"github.com/bb-consent/api/internal/policy"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/token"
	"github.com/bb-consent/api/internal/user"
)

type orgUpdateReq struct {
	Name        string `json:"name"`
	Location    string `json:"location"`
	Description string `json:"description"`
	PolicyURL   string `json:"policyUrl"`
}

type updateOrgResp struct {
	Organization organizationResp `json:"organisation"`
}

// updatePolicyUrl
func updatePolicyUrl(policyUrl string, organisationId string, orgAdminId string) error {
	// Repository
	prepo := policy.PolicyRepository{}
	prepo.Init(organisationId)

	toBeUpdatedPolicy, err := prepo.GetFirstPolicy()
	if err != nil {
		return err
	}

	toBeUpdatedPolicy.Url = policyUrl

	// Bump major version for policy
	updatedVersion, err := common.BumpMajorVersion(toBeUpdatedPolicy.Version)
	if err != nil {
		return err
	}
	toBeUpdatedPolicy.Version = updatedVersion

	_, err = prepo.Update(toBeUpdatedPolicy)
	if err != nil {
		return err
	}

	// Update revision
	_, err = revision.UpdateRevisionForPolicy(toBeUpdatedPolicy, orgAdminId)
	if err != nil {
		return err
	}
	return err
}

// UpdateOrganization Updates an organization
func UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	// Current user
	orgAdminId := token.GetUserID(r)

	var orgUpReq orgUpdateReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &orgUpReq)

	organizationID := common.Sanitize(r.Header.Get(config.OrganizationId))

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	if strings.TrimSpace(orgUpReq.Name) != "" {
		o.Name = orgUpReq.Name
	}
	if strings.TrimSpace(orgUpReq.Location) != "" {
		o.Location = orgUpReq.Location
	}
	if strings.TrimSpace(orgUpReq.Description) != "" {
		o.Description = orgUpReq.Description
	}
	if strings.TrimSpace(orgUpReq.PolicyURL) != "" {
		o.PolicyURL = orgUpReq.PolicyURL
	}

	orgResp, err := org.Update(o)
	if err != nil {
		m := fmt.Sprintf("Failed to update organization: %v", organizationID)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	go user.UpdateOrganizationsSubscribedUsers(orgResp)

	// update policy url
	err = updatePolicyUrl(orgResp.PolicyURL, organizationID, orgAdminId)
	if err != nil {
		m := fmt.Sprintf("Failed to update global policy url: %v", organizationID)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	oResp := organizationResp{
		ID:          orgResp.ID,
		Name:        orgResp.Name,
		Description: orgResp.Description,
		Sector:      orgResp.Type.Type,
		Location:    orgResp.Location,
		PolicyURL:   orgResp.PolicyURL,
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	response, _ := json.Marshal(updateOrgResp{oResp})
	w.Write(response)
}
