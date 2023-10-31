package dataagreement

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/org"
	"github.com/bb-consent/api/internal/policy"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/token"
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

type dataAttributeForDataAgreement struct {
	dataagreement.DataAttribute
	Id          string `json:"id"`
	Name        string `json:"name" validate:"required_if=Active true"`
	Description string `json:"description" validate:"required_if=Active true,max=500"`
}

type signatureForDataAgreement struct {
	dataagreement.Signature
	Id string `json:"id"`
}

type dataAgreement struct {
	Id                      primitive.ObjectID              `json:"id" bson:"_id,omitempty"`
	Version                 string                          `json:"version"`
	ControllerId            string                          `json:"controllerId"`
	ControllerUrl           string                          `json:"controllerUrl" validate:"required_if=Active true"`
	ControllerName          string                          `json:"controllerName" validate:"required_if=Active true"`
	Policy                  policyForDataAgreement          `json:"policy" validate:"required_if=Active true"`
	Purpose                 string                          `json:"purpose" validate:"required_if=Active true"`
	PurposeDescription      string                          `json:"purposeDescription" validate:"required_if=Active true,max=500"`
	LawfulBasis             string                          `json:"lawfulBasis" validate:"required_if=Active true"`
	MethodOfUse             string                          `json:"methodOfUse" validate:"required_if=Active true"`
	DpiaDate                string                          `json:"dpiaDate"`
	DpiaSummaryUrl          string                          `json:"dpiaSummaryUrl"`
	Signature               signatureForDataAgreement       `json:"signature"`
	Active                  bool                            `json:"active"`
	Forgettable             bool                            `json:"forgettable"`
	CompatibleWithVersionId string                          `json:"compatibleWithVersionId"`
	Lifecycle               string                          `json:"lifecycle" validate:"required_if=Active true"`
	DataAttributes          []dataAttributeForDataAgreement `json:"dataAttributes" validate:"required_if=Active true"`
	OrganisationId          string                          `json:"-"`
	IsDeleted               bool                            `json:"-"`
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
	fmt.Println(dataAgreementReq.DataAgreement.Active)

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

	newDataAgreement.Policy.Id = primitive.NewObjectID()
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
	newDataAgreement.Signature.Id = primitive.NewObjectID()
	newDataAgreement.Signature.Payload = requestBody.DataAgreement.Signature.Payload
	newDataAgreement.Signature.Signature = requestBody.DataAgreement.Signature.Signature.Signature
	newDataAgreement.Signature.VerificationMethod = requestBody.DataAgreement.Signature.VerificationMethod
	newDataAgreement.Signature.VerificationPayload = requestBody.DataAgreement.Signature.VerificationPayload
	newDataAgreement.Signature.VerificationPayloadHash = requestBody.DataAgreement.Signature.VerificationPayloadHash
	newDataAgreement.Signature.VerificationArtifact = requestBody.DataAgreement.Signature.VerificationArtifact
	newDataAgreement.Signature.VerificationSignedBy = requestBody.DataAgreement.Signature.VerificationSignedBy
	newDataAgreement.Signature.VerificationSignedAs = requestBody.DataAgreement.Signature.VerificationSignedAs
	newDataAgreement.Signature.VerificationJwsHeader = requestBody.DataAgreement.Signature.VerificationJwsHeader
	newDataAgreement.Signature.Timestamp = requestBody.DataAgreement.Signature.Timestamp
	newDataAgreement.Signature.SignedWithoutObjectReference = requestBody.DataAgreement.Signature.SignedWithoutObjectReference
	newDataAgreement.Signature.ObjectType = requestBody.DataAgreement.Signature.ObjectType
	newDataAgreement.Signature.ObjectReference = requestBody.DataAgreement.Signature.ObjectReference

	newDataAgreement.Active = requestBody.DataAgreement.Active
	newDataAgreement.Forgettable = requestBody.DataAgreement.Forgettable
	newDataAgreement.CompatibleWithVersionId = requestBody.DataAgreement.CompatibleWithVersionId

	return newDataAgreement
}

func updateDataAttributeFromAddDataAgreementRequestBody(requestBody addDataAgreementReq) []dataagreement.DataAttribute {
	var newDataAttributes []dataagreement.DataAttribute

	for _, dA := range requestBody.DataAgreement.DataAttributes {
		var dataAttribute dataagreement.DataAttribute
		dataAttribute.Id = primitive.NewObjectID()
		dataAttribute.Name = dA.Name
		dataAttribute.Description = dA.Description
		dataAttribute.Category = dA.Category
		dataAttribute.Sensitivity = dA.Sensitivity

		newDataAttributes = append(newDataAttributes, dataAttribute)
	}

	return newDataAttributes
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
	newDataAttributes := updateDataAttributeFromAddDataAgreementRequestBody(dataAgreementReq)
	// Update data agreement from request body
	newDataAgreement = updateDataAgreementFromAddDataAgreementRequestBody(dataAgreementReq, newDataAgreement)
	newDataAgreement.OrganisationId = organisationId
	newDataAgreement.ControllerId = organisationId
	newDataAgreement.ControllerName = o.Name
	newDataAgreement.ControllerUrl = o.EulaURL
	newDataAgreement.DataAttributes = newDataAttributes
	newDataAgreement.IsDeleted = false
	newDataAgreement.Lifecycle = lifecycle

	var newRevision revision.Revision
	if lifecycle == config.Complete {
		newDataAgreement.Version = version

		// Create new revision
		newRevision, err = revision.CreateRevisionForDataAgreement(newDataAgreement, orgAdminId)
		if err != nil {
			m := fmt.Sprintf("Failed to create revision for new data agreement: %v", newDataAgreement.Id)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

		// Save the data agreement to db
		_, err := revision.Add(newRevision)
		if err != nil {
			m := fmt.Sprintf("Failed to create new revision: %v", newRevision.Id)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
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

	// Constructing the response
	var resp addDataAgreementResp
	resp.DataAgreement = savedDataAgreement

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(newRevision)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}