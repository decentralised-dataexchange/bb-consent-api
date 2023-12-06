package dataagreement

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/org"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/token"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type updateDataAgreementReq struct {
	DataAgreement dataAgreement `json:"dataAgreement"`
}

type updateDataAgreementResp struct {
	DataAgreement dataagreement.DataAgreement `json:"dataAgreement"`
	Revision      interface{}                 `json:"revision"`
}

func validateUpdateDataAgreementRequestBody(dataAgreementReq updateDataAgreementReq) error {
	// validating request payload
	var validate = validator.New()

	if err := validate.Struct(dataAgreementReq.DataAgreement); err != nil {
		return err
	}

	// Proceed if lawful basis provided is valid
	if !isValidLawfulBasisOfProcessing(dataAgreementReq.DataAgreement.LawfulBasis) {
		return errors.New("invalid lawful basis provided")
	}

	if len(strings.TrimSpace(dataAgreementReq.DataAgreement.DataUse)) < 1 && len(strings.TrimSpace(dataAgreementReq.DataAgreement.MethodOfUse)) < 1 {
		return errors.New("missing mandatory param dataUse")
	}

	if len(strings.TrimSpace(dataAgreementReq.DataAgreement.MethodOfUse)) > 1 && !isValidMethodOfUse(dataAgreementReq.DataAgreement.MethodOfUse) {
		return errors.New("invalid method of use provided")
	}

	if len(strings.TrimSpace(dataAgreementReq.DataAgreement.DataUse)) > 1 && !isValidMethodOfUse(dataAgreementReq.DataAgreement.DataUse) {
		return errors.New("invalid data use provided")
	}

	return nil
}

func dataAttributeIdExists(dataAttributeId string, currentDataAttributes []dataagreement.DataAttribute) bool {
	for _, dataAttribute := range currentDataAttributes {
		if dataAttributeId == dataAttribute.Id {
			return true
		}
	}
	return false
}

func updateDataAttributeFromUpdateDataAgreementRequestBody(requestBody updateDataAgreementReq, currentDataAttributes []dataagreement.DataAttribute) []dataagreement.DataAttribute {
	var newDataAttributes []dataagreement.DataAttribute

	for _, dA := range requestBody.DataAgreement.DataAttributes {
		var dataAttribute dataagreement.DataAttribute

		isExistingDataAttribute := dataAttributeIdExists(dA.Id, currentDataAttributes)
		if isExistingDataAttribute {
			dataAttribute.Id = dA.Id
		} else {
			dataAttribute.Id = primitive.NewObjectID().Hex()
		}

		dataAttribute.Name = dA.Name
		dataAttribute.Description = dA.Description
		dataAttribute.Category = dA.Category
		dataAttribute.Sensitivity = dA.Sensitivity

		newDataAttributes = append(newDataAttributes, dataAttribute)
	}

	return newDataAttributes
}

func updateControllerFromReq(o org.Organization, toBeUpdatedDataAgreement dataagreement.DataAgreement) dataagreement.DataAgreement {
	toBeUpdatedDataAgreement.ControllerId = o.ID
	toBeUpdatedDataAgreement.ControllerName = o.Name
	toBeUpdatedDataAgreement.ControllerUrl = o.EulaURL

	toBeUpdatedDataAgreement.Controller.Id = o.ID
	toBeUpdatedDataAgreement.Controller.Name = o.Name
	toBeUpdatedDataAgreement.Controller.Url = o.EulaURL
	return toBeUpdatedDataAgreement
}

func updateDataAgreementFromRequestBody(requestBody updateDataAgreementReq, toBeUpdatedDataAgreement dataagreement.DataAgreement) dataagreement.DataAgreement {

	toBeUpdatedDataAgreement.Policy.Name = requestBody.DataAgreement.Policy.Name
	toBeUpdatedDataAgreement.Policy.Version = requestBody.DataAgreement.Policy.Version
	toBeUpdatedDataAgreement.Policy.Url = requestBody.DataAgreement.Policy.Url
	toBeUpdatedDataAgreement.Policy.Jurisdiction = requestBody.DataAgreement.Policy.Jurisdiction
	toBeUpdatedDataAgreement.Policy.IndustrySector = requestBody.DataAgreement.Policy.IndustrySector
	toBeUpdatedDataAgreement.Policy.DataRetentionPeriodDays = requestBody.DataAgreement.Policy.DataRetentionPeriodDays
	toBeUpdatedDataAgreement.Policy.GeographicRestriction = requestBody.DataAgreement.Policy.GeographicRestriction
	toBeUpdatedDataAgreement.Policy.StorageLocation = requestBody.DataAgreement.Policy.StorageLocation
	toBeUpdatedDataAgreement.Policy.ThirdPartyDataSharing = requestBody.DataAgreement.Policy.ThirdPartyDataSharing

	toBeUpdatedDataAgreement.Purpose = requestBody.DataAgreement.Purpose
	toBeUpdatedDataAgreement.PurposeDescription = requestBody.DataAgreement.PurposeDescription
	toBeUpdatedDataAgreement.LawfulBasis = requestBody.DataAgreement.LawfulBasis
	toBeUpdatedDataAgreement.MethodOfUse = requestBody.DataAgreement.MethodOfUse
	toBeUpdatedDataAgreement.DpiaDate = requestBody.DataAgreement.DpiaDate
	toBeUpdatedDataAgreement.DpiaSummaryUrl = requestBody.DataAgreement.DpiaSummaryUrl
	toBeUpdatedDataAgreement.Dpia = requestBody.DataAgreement.Dpia
	toBeUpdatedDataAgreement.CompatibleWithVersion = requestBody.DataAgreement.CompatibleWithVersion

	toBeUpdatedDataAgreement.Signature.Payload = requestBody.DataAgreement.Signature.Payload
	toBeUpdatedDataAgreement.Signature.Signature = requestBody.DataAgreement.Signature.Signature
	toBeUpdatedDataAgreement.Signature.VerificationMethod = requestBody.DataAgreement.Signature.VerificationMethod
	toBeUpdatedDataAgreement.Signature.VerificationPayload = requestBody.DataAgreement.Signature.VerificationPayload
	toBeUpdatedDataAgreement.Signature.VerificationPayloadHash = requestBody.DataAgreement.Signature.VerificationPayloadHash
	toBeUpdatedDataAgreement.Signature.VerificationArtifact = requestBody.DataAgreement.Signature.VerificationArtifact
	toBeUpdatedDataAgreement.Signature.VerificationSignedBy = requestBody.DataAgreement.Signature.VerificationSignedBy
	toBeUpdatedDataAgreement.Signature.VerificationSignedAs = requestBody.DataAgreement.Signature.VerificationSignedAs
	toBeUpdatedDataAgreement.Signature.VerificationJwsHeader = requestBody.DataAgreement.Signature.VerificationJwsHeader
	toBeUpdatedDataAgreement.Signature.Timestamp = requestBody.DataAgreement.Signature.Timestamp
	toBeUpdatedDataAgreement.Signature.SignedWithoutObjectReference = requestBody.DataAgreement.Signature.SignedWithoutObjectReference
	toBeUpdatedDataAgreement.Signature.ObjectType = requestBody.DataAgreement.Signature.ObjectType
	toBeUpdatedDataAgreement.Signature.ObjectReference = requestBody.DataAgreement.Signature.ObjectReference

	toBeUpdatedDataAgreement.Active = requestBody.DataAgreement.Active
	toBeUpdatedDataAgreement.Forgettable = requestBody.DataAgreement.Forgettable
	toBeUpdatedDataAgreement.CompatibleWithVersionId = requestBody.DataAgreement.CompatibleWithVersionId

	// Update life cycle based on active field
	toBeUpdatedDataAgreement.Lifecycle = setDataAgreementLifecycle(requestBody.DataAgreement.Active)

	dataAttributes := updateDataAttributeFromUpdateDataAgreementRequestBody(requestBody, toBeUpdatedDataAgreement.DataAttributes)

	toBeUpdatedDataAgreement.DataAttributes = dataAttributes

	// update method of use if data use not empty and is valid
	if len(strings.TrimSpace(requestBody.DataAgreement.DataUse)) > 0 && isValidMethodOfUse(requestBody.DataAgreement.DataUse) {
		toBeUpdatedDataAgreement.DataUse = requestBody.DataAgreement.DataUse
		toBeUpdatedDataAgreement.MethodOfUse = requestBody.DataAgreement.DataUse
	} else {
		toBeUpdatedDataAgreement.DataUse = requestBody.DataAgreement.MethodOfUse
	}

	return toBeUpdatedDataAgreement
}

// ConfigUpdateDataAgreement
func ConfigUpdateDataAgreement(w http.ResponseWriter, r *http.Request) {
	// Current user
	orgAdminId := token.GetUserID(r)

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Path params
	dataAgreementId := mux.Vars(r)[config.DataAgreementId]
	dataAgreementId = common.Sanitize(dataAgreementId)

	// Request body
	var dataAgreementReq updateDataAgreementReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &dataAgreementReq)

	// Validate request body
	err := validateUpdateDataAgreementRequestBody(dataAgreementReq)
	if err != nil {
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Query organisation by Id
	o, err := org.Get(organisationId)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization by ID :%v", organisationId)
		common.HandleErrorV2(w, http.StatusNotFound, m, err)
		return
	}

	// Repository
	daRepo := dataagreement.DataAgreementRepository{}
	daRepo.Init(organisationId)

	count, err := daRepo.CountDocumentsByPurposeExeptOneDataAgreement(strings.TrimSpace(dataAgreementReq.DataAgreement.Purpose), dataAgreementId)
	if err != nil {
		m := "Failed to count data agreements by purpose"
		common.HandleErrorV2(w, http.StatusNotFound, m, err)
		return
	}
	if count >= 1 {
		m := "Data agreement purpose exists"
		common.HandleErrorV2(w, http.StatusBadRequest, m, err)
		return
	}

	daRepo.Init(organisationId)

	// Get data agreement from db
	currentDataAgreement, err := daRepo.Get(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement by id: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	currentVersion := currentDataAgreement.Version

	// Update data agreement from request body
	toBeUpdatedDataAgreement := updateDataAgreementFromRequestBody(dataAgreementReq, currentDataAgreement)
	toBeUpdatedDataAgreement = updateControllerFromReq(o, toBeUpdatedDataAgreement)

	// Bump major version for data agreement
	updatedVersion, err := common.BumpMajorVersion(toBeUpdatedDataAgreement.Version)
	if err != nil {
		m := fmt.Sprintf("Failed to bump major version for data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Increment data agreement version
	toBeUpdatedDataAgreement.Version = updatedVersion

	var newRevision revision.Revision

	// if data agreement is in draft mode then
	// version is not incremented
	if !currentDataAgreement.Active {
		toBeUpdatedDataAgreement.Version = currentVersion
	}

	// If data agreement is published then:
	//a. Add a new revision
	if toBeUpdatedDataAgreement.Active {

		// Update revision
		newRevision, err = revision.UpdateRevisionForDataAgreement(toBeUpdatedDataAgreement, orgAdminId)
		if err != nil {
			m := fmt.Sprintf("Failed to update data agreement: %v", dataAgreementId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

	} else {
		// If data agreement is draft then:
		// a. Create a revision on runtime
		newRevision, err = revision.CreateRevisionForDraftDataAgreement(toBeUpdatedDataAgreement, orgAdminId)
		if err != nil {
			m := fmt.Sprintf("Failed to create revision for draft data agreement: %v", dataAgreementId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
	}

	// Save the data agreement to db
	savedDataAgreement, err := daRepo.Update(toBeUpdatedDataAgreement)
	if err != nil {
		m := fmt.Sprintf("Failed to update data agreement: %v", toBeUpdatedDataAgreement.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Constructing the response
	var resp updateDataAgreementResp
	resp.DataAgreement = savedDataAgreement

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(newRevision)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
