package dataagreement

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/dataagreement"
	"github.com/bb-consent/api/src/revision"
	"github.com/bb-consent/api/src/token"
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

type addDataAgreementReq struct {
	DataAgreement dataagreement.DataAgreement `json:"dataAgreement" valid:"required"`
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

func validateAddDataAgreementRequestBody(dataAgreementReq addDataAgreementReq) error {
	// validating request payload
	valid, err := govalidator.ValidateStruct(dataAgreementReq)
	if !valid {
		return err
	}

	// Proceed if lawful basis provided is valid
	if !isValidLawfulBasisOfProcessing(dataAgreementReq.DataAgreement.LawfulBasis) {
		return errors.New("invalid lawful basis provided")
	}

	return nil
}

func updateDataAgreementFromAddDataAgreementRequestBody(requestBody addDataAgreementReq, newDataAgreement dataagreement.DataAgreement) dataagreement.DataAgreement {

	newDataAgreement.ControllerUrl = requestBody.DataAgreement.ControllerUrl
	newDataAgreement.ControllerName = requestBody.DataAgreement.ControllerName
	newDataAgreement.Policy = requestBody.DataAgreement.Policy
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
	newDataAgreement.Lifecycle = requestBody.DataAgreement.Lifecycle

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

	version := common.IntegerToSemver(1)

	// Initialise data agreement
	var newDataAgreement dataagreement.DataAgreement
	newDataAgreement.Id = primitive.NewObjectID().Hex()
	// Update data agreement from request body
	newDataAgreement = updateDataAgreementFromAddDataAgreementRequestBody(dataAgreementReq, newDataAgreement)
	newDataAgreement.OrganisationId = organisationId
	newDataAgreement.ControllerId = organisationId
	newDataAgreement.IsDeleted = false
	newDataAgreement.Version = version

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
