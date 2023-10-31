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

	// Proceed if method of use is valid
	if !isValidMethodOfUse(dataAgreementReq.DataAgreement.MethodOfUse) {
		return errors.New("invalid method of use provided")
	}

	return nil
}

func updateDataAttributeFromUpdateDataAgreementRequestBody(requestBody updateDataAgreementReq) []dataagreement.DataAttribute {
	var newDataAttributes []dataagreement.DataAttribute

	for _, dA := range requestBody.DataAgreement.DataAttributes {
		var dataAttribute dataagreement.DataAttribute
		if dA.DataAttribute.Id.IsZero() {
			dataAttribute.Id = primitive.NewObjectID()
		} else {
			dataAttribute.Id = dA.DataAttribute.Id
		}
		dataAttribute.Name = dA.Name
		dataAttribute.Description = dA.Description
		dataAttribute.Category = dA.Category
		dataAttribute.Sensitivity = dA.Sensitivity

		newDataAttributes = append(newDataAttributes, dataAttribute)
	}

	return newDataAttributes
}

func updateDataAgreementFromRequestBody(requestBody updateDataAgreementReq, toBeUpdatedDataAgreement dataagreement.DataAgreement) dataagreement.DataAgreement {

	toBeUpdatedDataAgreement.Policy.Id = requestBody.DataAgreement.Policy.Policy.Id
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

	toBeUpdatedDataAgreement.Signature.Id = requestBody.DataAgreement.Signature.Signature.Id
	toBeUpdatedDataAgreement.Signature.Payload = requestBody.DataAgreement.Signature.Payload
	toBeUpdatedDataAgreement.Signature.Signature = requestBody.DataAgreement.Signature.Signature.Signature
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

	if requestBody.DataAgreement.Signature.Signature.Id.IsZero() {
		toBeUpdatedDataAgreement.Signature.Id = primitive.NewObjectID()
	}

	if requestBody.DataAgreement.Policy.Policy.Id.IsZero() {
		toBeUpdatedDataAgreement.Policy.Id = primitive.NewObjectID()
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
	o, err := org.Get(organisationId)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization by ID :%v", organisationId)
		common.HandleErrorV2(w, http.StatusNotFound, m, err)
		return
	}

	// Repository
	daRepo := dataagreement.DataAgreementRepository{}
	daRepo.Init(organisationId)
	// Get policy from db
	toBeUpdatedDataAgreement, err := daRepo.Get(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement by id: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Update data agreement from request body
	toBeUpdatedDataAgreement = updateDataAgreementFromRequestBody(dataAgreementReq, toBeUpdatedDataAgreement)

	// Update life cycle based on active field
	lifecycle := setDataAgreementLifecycle(dataAgreementReq.DataAgreement.Active)

	toBeUpdatedDataAgreement.ControllerName = o.Name
	toBeUpdatedDataAgreement.ControllerUrl = o.EulaURL
	toBeUpdatedDataAgreement.Lifecycle = lifecycle

	toBeUpdatedDataAttributes := updateDataAttributeFromUpdateDataAgreementRequestBody(dataAgreementReq)
	toBeUpdatedDataAgreement.DataAttributes = toBeUpdatedDataAttributes

	// Bump major version for policy
	updatedVersion, err := common.BumpMajorVersion(toBeUpdatedDataAgreement.Version)
	if err != nil {
		m := fmt.Sprintf("Failed to bump major version for data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	var currentRevision revision.Revision
	var newRevision revision.Revision

	if toBeUpdatedDataAgreement.Active {

		if toBeUpdatedDataAgreement.Version == "0.0.0" {
			toBeUpdatedDataAgreement.Version = updatedVersion
			// Create new revision
			newRevision, err = revision.CreateRevisionForDataAgreement(toBeUpdatedDataAgreement, orgAdminId)
			if err != nil {
				m := fmt.Sprintf("Failed to create revision for data agreement: %v", toBeUpdatedDataAgreement.Id)
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}
		} else {
			// Get revision from db
			currentRevision, err = revision.GetLatestByDataAgreementId(dataAgreementId)
			if err != nil {
				m := fmt.Sprintf("Failed to fetch latest revision by data agreement id: %v", dataAgreementId)
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}

			toBeUpdatedDataAgreement.Version = updatedVersion

			// Update revision
			newRevision, err = revision.UpdateRevisionForDataAgreement(toBeUpdatedDataAgreement, &currentRevision, orgAdminId)
			if err != nil {
				m := fmt.Sprintf("Failed to update revision for data agreement: %v", dataAgreementId)
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}

			// Save the previous revision to db
			updatedRevision, err := revision.Update(currentRevision)
			if err != nil {
				m := fmt.Sprintf("Failed to update revision: %v", updatedRevision.Id)
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}
		}

		// Save the revision to db
		currentRevision, err = revision.Add(newRevision)
		if err != nil {
			m := fmt.Sprintf("Failed to create new data agreement: %v", newRevision.Id)
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
	revisionForHTTPResponse.Init(currentRevision)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
