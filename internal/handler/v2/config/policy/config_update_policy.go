package policy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/policy"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/token"
	"github.com/gorilla/mux"
)

type updatePolicyReq struct {
	Policy policy.Policy `json:"policy" valid:"required"`
}

type updatePolicyResp struct {
	Policy   policy.Policy `json:"policy"`
	Revision interface{}   `json:"revision"`
}

func validateUpdatePolicyRequestBody(policyReq updatePolicyReq) error {
	// validating request payload
	valid, err := govalidator.ValidateStruct(policyReq)
	if err != nil {
		return err
	}

	if !valid {
		return errors.New("invalid request payload")
	}

	if strings.TrimSpace(policyReq.Policy.Name) == "" {
		return errors.New("missing mandatory param - Name")
	}

	if strings.TrimSpace(policyReq.Policy.Url) == "" {
		return errors.New("missing mandatory param - Url")

	}

	return nil
}

func updatePolicyFromRequestBody(requestBody updatePolicyReq, toBeUpdatedPolicy policy.Policy) policy.Policy {
	toBeUpdatedPolicy.Name = requestBody.Policy.Name
	toBeUpdatedPolicy.Url = requestBody.Policy.Url
	toBeUpdatedPolicy.Jurisdiction = requestBody.Policy.Jurisdiction
	toBeUpdatedPolicy.IndustrySector = requestBody.Policy.IndustrySector
	toBeUpdatedPolicy.DataRetentionPeriodDays = requestBody.Policy.DataRetentionPeriodDays
	toBeUpdatedPolicy.GeographicRestriction = requestBody.Policy.GeographicRestriction
	toBeUpdatedPolicy.StorageLocation = requestBody.Policy.StorageLocation
	toBeUpdatedPolicy.ThirdPartyDataSharing = requestBody.Policy.ThirdPartyDataSharing
	return toBeUpdatedPolicy
}

// ConfigUpdatePolicy
func ConfigUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	// Current user
	orgAdminId := token.GetUserID(r)

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Path params
	policyId := mux.Vars(r)[config.PolicyId]
	policyId = common.Sanitize(policyId)

	// Request body
	var policyReq updatePolicyReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &policyReq)

	// Validate request body
	err := validateUpdatePolicyRequestBody(policyReq)
	if err != nil {
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Repository
	policyRepo := policy.PolicyRepository{}
	policyRepo.Init(organisationId)

	// Get policy from db
	toBeUpdatedPolicy, err := policyRepo.Get(policyId)
	if err != nil {
		common.HandleErrorV2(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	// Update policy from request body
	toBeUpdatedPolicy = updatePolicyFromRequestBody(policyReq, toBeUpdatedPolicy)

	// Bump major version for policy
	updatedVersion, err := common.BumpMajorVersion(toBeUpdatedPolicy.Version)
	if err != nil {
		m := fmt.Sprintf("Failed to bump major version for policy: %v", policyId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	toBeUpdatedPolicy.Version = updatedVersion

	// Update revision
	newRevision, err := revision.UpdateRevisionForPolicy(toBeUpdatedPolicy, orgAdminId)
	if err != nil {
		m := fmt.Sprintf("Failed to update revision for policy: %v", policyId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Save the policy to db
	savedPolicy, err := policyRepo.Update(toBeUpdatedPolicy)
	if err != nil {
		m := fmt.Sprintf("Failed to update policy: %v", toBeUpdatedPolicy.Name)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	policyRepo.Init(organisationId)
	first_policy, err := policyRepo.GetFirstPolicy()
	if err != nil {
		m := "Failed to fetch first policy"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	if first_policy.Id == toBeUpdatedPolicy.Id {
		// updates organisation policy url
		err = updateOrganisationPolicyUrl(savedPolicy.Url, organisationId)
		if err != nil {
			m := fmt.Sprintf("Failed to update organisation policy url: %v", organisationId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
	}

	// Constructing the response
	var resp updatePolicyResp
	resp.Policy = savedPolicy

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(newRevision)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
