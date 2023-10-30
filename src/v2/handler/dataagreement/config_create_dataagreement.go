package dataagreement

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/policy"
	"github.com/bb-consent/api/src/v2/dataagreement"
	"github.com/bb-consent/api/src/v2/revision"
	"github.com/bb-consent/api/src/v2/token"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// LawfulBasisOfProcessingMapping Structure defining lawful basis of processing label
type LawfulBasisOfProcessingMapping struct {
	Str string
}

// LawfulBasisOfProcessingMappings List of available lawful basis of processing mappings
var LawfulBasisOfProcessingMappings = []LawfulBasisOfProcessingMapping{
	{
		"consent",
	},
	{
		"contract",
	},
	{

		"legal_obligation",
	},
	{
		"vital_interest",
	},
	{
		"public_task",
	},
	{
		"legitimate_interest",
	},
}

type MethodOfUseMapping struct {
	Str string
}

// MethodOfUseMappings List of available method of use
var MethodOfUseMappings = []MethodOfUseMapping{
	{
		"null",
	},
	{
		"data_source",
	},
	{

		"data_using_service",
	},
}

type policyForDataAgreement struct {
	policy.Policy
	Id   string `json:"id"`
	Name string `json:"name" validate:"required_if=Active true"`
	Url  string `json:"url" validate:"required_if=Active true"`
}

type dataAgreement struct {
	Id                      primitive.ObjectID      `json:"id" bson:"_id,omitempty"`
	Version                 string                  `json:"version"`
	ControllerId            string                  `json:"controllerId"`
	ControllerUrl           string                  `json:"controllerUrl" validate:"required_if=Active true"`
	ControllerName          string                  `json:"controllerName" validate:"required_if=Active true"`
	Policy                  policyForDataAgreement  `json:"policy" validate:"required_if=Active true"`
	Purpose                 string                  `json:"purpose" validate:"required_if=Active true"`
	PurposeDescription      string                  `json:"purposeDescription" validate:"required_if=Active true,max=500"`
	LawfulBasis             string                  `json:"lawfulBasis" validate:"required_if=Active true"`
	MethodOfUse             string                  `json:"methodOfUse" validate:"required_if=Active true"`
	DpiaDate                string                  `json:"dpiaDate"`
	DpiaSummaryUrl          string                  `json:"dpiaSummaryUrl"`
	Signature               dataagreement.Signature `json:"signature" validate:"required_if=Active true"`
	Active                  bool                    `json:"active"`
	Forgettable             bool                    `json:"forgettable"`
	CompatibleWithVersionId string                  `json:"compatibleWithVersionId"`
	Lifecycle               string                  `json:"lifecycle" validate:"required_if=Active true"`
	OrganisationId          string                  `json:"-"`
	IsDeleted               bool                    `json:"-"`
}

type addDataAgreementReq struct {
	DataAgreement dataAgreement `json:"dataAgreement"`
}

type addDataAgreementResp struct {
	DataAgreement dataagreement.DataAgreement `json:"dataAgreement"`
	Revision      interface{}                 `json:"revision"`
}

// Check if the lawful usage ID provided is valid
func isValidLawfulBasisOfProcessing(lawfulBasis string) bool {
	isFound := false
	for _, lawfulBasisOfProcessingMapping := range LawfulBasisOfProcessingMappings {
		if lawfulBasisOfProcessingMapping.Str == lawfulBasis {
			isFound = true
			break
		}
	}

	return isFound
}

// Check if the method of use provided is valid
func isValidMethodOfUse(methodOfUse string) bool {
	isFound := false
	for _, MethodOfUseMapping := range MethodOfUseMappings {
		if MethodOfUseMapping.Str == methodOfUse {
			isFound = true
			break
		}
	}

	return isFound
}

func validateAddDataAgreementRequestBody(dataAgreementReq addDataAgreementReq) error {
	var validate = validator.New()

	if err := validate.Struct(dataAgreementReq.DataAgreement); err != nil {
		return err
	}

	if dataAgreementReq.DataAgreement.Active {
		// Proceed if lawful basis provided is valid
		if !isValidLawfulBasisOfProcessing(dataAgreementReq.DataAgreement.LawfulBasis) {
			return errors.New("invalid lawful basis provided")
		}

		// Proceed if method of use is valid
		if !isValidMethodOfUse(dataAgreementReq.DataAgreement.MethodOfUse) {
			return errors.New("invalid method of use provided")
		}
	}

	return nil
}

func setDataAgreementLifecycle(active bool) string {
	var lifecycle string
	if active {
		lifecycle = "complete"
	} else {
		lifecycle = "draft"
	}
	return lifecycle
}

func updateDataAgreementFromAddDataAgreementRequestBody(requestBody addDataAgreementReq, newDataAgreement dataagreement.DataAgreement) dataagreement.DataAgreement {

	newDataAgreement.Policy.Id = requestBody.DataAgreement.Policy.Id
	newDataAgreement.Policy.Name = requestBody.DataAgreement.Policy.Name
	newDataAgreement.Policy.Version = requestBody.DataAgreement.Policy.Version
	newDataAgreement.Policy.Url = requestBody.DataAgreement.Policy.Url
	newDataAgreement.Policy.Jurisdiction = requestBody.DataAgreement.Policy.Jurisdiction
	newDataAgreement.Policy.IndustrySector = requestBody.DataAgreement.Policy.IndustrySector
	newDataAgreement.Policy.DataRetentionPeriodDays = requestBody.DataAgreement.Policy.DataRetentionPeriodDays
	newDataAgreement.Policy.GeographicRestriction = requestBody.DataAgreement.Policy.GeographicRestriction
	newDataAgreement.Policy.StorageLocation = requestBody.DataAgreement.Policy.StorageLocation
	newDataAgreement.Policy.ThirdPartyDataSharing = requestBody.DataAgreement.Policy.ThirdPartyDataSharing
	newDataAgreement.Purpose = requestBody.DataAgreement.Purpose
	newDataAgreement.PurposeDescription = requestBody.DataAgreement.PurposeDescription
	newDataAgreement.LawfulBasis = requestBody.DataAgreement.LawfulBasis
	newDataAgreement.MethodOfUse = requestBody.DataAgreement.MethodOfUse
	newDataAgreement.DpiaDate = requestBody.DataAgreement.DpiaDate
	newDataAgreement.DpiaSummaryUrl = requestBody.DataAgreement.DpiaSummaryUrl
	newDataAgreement.Signature = requestBody.DataAgreement.Signature
	newDataAgreement.Active = requestBody.DataAgreement.Active
	newDataAgreement.Forgettable = requestBody.DataAgreement.Forgettable
	newDataAgreement.CompatibleWithVersionId = requestBody.DataAgreement.CompatibleWithVersionId

	return newDataAgreement
}

// ConfigCreatePolicy
func ConfigCreateDataAgreement(w http.ResponseWriter, r *http.Request) {
	// Current user
	orgAdminId := token.GetUserID(r)

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Request body
	var dataAgreementReq addDataAgreementReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &dataAgreementReq)

	// Validate request body
	err := validateAddDataAgreementRequestBody(dataAgreementReq)
	if err != nil {
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	o, err := org.Get(organisationId)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization by ID :%v", organisationId)
		common.HandleErrorV2(w, http.StatusNotFound, m, err)
		return
	}

	version := common.IntegerToSemver(1)
	// Add life cycle based on active field
	lifecycle := setDataAgreementLifecycle(dataAgreementReq.DataAgreement.Active)

	// Initialise data agreement
	var newDataAgreement dataagreement.DataAgreement
	newDataAgreement.Id = primitive.NewObjectID()
	// Update data agreement from request body
	newDataAgreement = updateDataAgreementFromAddDataAgreementRequestBody(dataAgreementReq, newDataAgreement)
	newDataAgreement.OrganisationId = organisationId
	newDataAgreement.ControllerId = organisationId
	newDataAgreement.ControllerName = o.Name
	newDataAgreement.ControllerUrl = o.EulaURL
	newDataAgreement.IsDeleted = false
	if lifecycle == config.Complete {
		newDataAgreement.Version = version
	}
	newDataAgreement.Lifecycle = lifecycle

	// Create new revision
	newRevision, err := revision.CreateRevisionForDataAgreement(newDataAgreement, orgAdminId)
	if err != nil {
		m := fmt.Sprintf("Failed to create revision for new data agreement: %v", newDataAgreement.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Repository
	darepo := dataagreement.DataAgreementRepository{}
	darepo.Init(organisationId)

	// Save the data agreement to db
	savedDataAgreement, err := darepo.Add(newDataAgreement)
	if err != nil {
		m := fmt.Sprintf("Failed to create new data agreement: %v", newDataAgreement.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Save the data agreement to db
	savedRevision, err := revision.Add(newRevision)
	if err != nil {
		m := fmt.Sprintf("Failed to create new revision: %v", newRevision.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Constructing the response
	var resp addDataAgreementResp
	resp.DataAgreement = savedDataAgreement

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(savedRevision)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
