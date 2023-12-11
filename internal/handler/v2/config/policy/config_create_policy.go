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
	"github.com/bb-consent/api/internal/org"
	"github.com/bb-consent/api/internal/policy"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/token"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RevisionForSnapshot struct {
	revision.Revision
	SerializedSnapshot   string `json:"-"`
	Id                   string `json:"-"`
	SuccessorId          string `json:"-"`
	PredecessorHash      string `json:"-"`
	PredecessorSignature string `json:"-"`
	SerializedHash       string `json:"-"`
}

type addPolicyReq struct {
	Policy policy.Policy `json:"policy" valid:"required"`
}

type addPolicyResp struct {
	Policy   policy.Policy `json:"policy"`
	Revision interface{}   `json:"revision"`
}

func validateAddPolicyRequestBody(policyReq addPolicyReq) error {
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

func updatePolicyFromAddPolicyRequestBody(requestBody addPolicyReq, newPolicy policy.Policy) policy.Policy {
	newPolicy.Name = requestBody.Policy.Name
	newPolicy.Url = requestBody.Policy.Url
	newPolicy.Jurisdiction = requestBody.Policy.Jurisdiction
	newPolicy.IndustrySector = requestBody.Policy.IndustrySector
	newPolicy.DataRetentionPeriodDays = requestBody.Policy.DataRetentionPeriodDays
	newPolicy.GeographicRestriction = requestBody.Policy.GeographicRestriction
	newPolicy.StorageLocation = requestBody.Policy.StorageLocation
	newPolicy.ThirdPartyDataSharing = requestBody.Policy.ThirdPartyDataSharing
	return newPolicy
}

// update organistaion policy url
func updateOrganisationPolicyUrl(policyUrl string, organisationId string) error {
	organisation, err := org.Get(organisationId)
	if err != nil {
		return err
	}
	organisation.PolicyURL = policyUrl
	_, err = org.Update(organisation)
	return err
}

// ConfigCreatePolicy
func ConfigCreatePolicy(w http.ResponseWriter, r *http.Request) {
	// Current user
	orgAdminId := token.GetUserID(r)

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Repository
	prepo := policy.PolicyRepository{}
	prepo.Init(organisationId)

	// Request body
	var policyReq addPolicyReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &policyReq)

	// Validate request body
	err := validateAddPolicyRequestBody(policyReq)
	if err != nil {
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	version := common.IntegerToSemver(1)

	// Initialise policy
	var newPolicy policy.Policy
	newPolicy.Id = primitive.NewObjectID().Hex()
	// Update policy from request body
	newPolicy = updatePolicyFromAddPolicyRequestBody(policyReq, newPolicy)
	newPolicy.OrganisationId = organisationId
	newPolicy.IsDeleted = false
	newPolicy.Version = version

	// Create new revision
	newRevision, err := revision.CreateRevisionForPolicy(newPolicy, orgAdminId)
	if err != nil {
		m := fmt.Sprintf("Failed to create revision for new policy: %v", newPolicy.Name)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Save the policy to db
	savedPolicy, err := prepo.Add(newPolicy)
	if err != nil {
		m := fmt.Sprintf("Failed to create new policy: %v", newPolicy.Name)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Save the revision to db
	savedRevision, err := revision.Add(newRevision)
	if err != nil {
		m := fmt.Sprintf("Failed to create new revision: %v", newRevision.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	count, err := prepo.GetPolicyCountByOrganisation()
	if err != nil {
		m := "Failed to count policies"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	if count == 1 {
		// updates organisation policy url
		err = updateOrganisationPolicyUrl(savedPolicy.Url, organisationId)
		if err != nil {
			m := fmt.Sprintf("Failed to update organisation policy url: %v", organisationId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
	}

	// Constructing the response
	var resp addPolicyResp
	resp.Policy = savedPolicy

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(savedRevision)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
