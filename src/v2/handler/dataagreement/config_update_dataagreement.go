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
	"github.com/gorilla/mux"
)

type updateDataAgreementReq struct {
	DataAgreement dataagreement.DataAgreement `json:"dataAgreement" valid:"required"`
}

type updateDataAgreementResp struct {
	DataAgreement dataagreement.DataAgreement `json:"dataAgreement"`
	Revision      interface{}                 `json:"revision"`
}

func validateUpdateDataAgreementRequestBody(dataAgreementReq updateDataAgreementReq) error {
	// validating request payload
	valid, err := govalidator.ValidateStruct(dataAgreementReq)
	if err != nil {
		return err
	}

	if !valid {
		return errors.New("invalid request payload")
	}

	// Proceed if lawful basis provided is valid
	if !isValidLawfulBasisOfProcessing(dataAgreementReq.DataAgreement.LawfulBasis) {
		return errors.New("invalid lawful basis provided")
	}

	return nil
}

func updateDataAgreementFromRequestBody(requestBody updateDataAgreementReq, toBeUpdatedDataAgreement dataagreement.DataAgreement) dataagreement.DataAgreement {

	toBeUpdatedDataAgreement.ControllerUrl = requestBody.DataAgreement.ControllerUrl
	toBeUpdatedDataAgreement.ControllerName = requestBody.DataAgreement.ControllerName
	toBeUpdatedDataAgreement.Policy = requestBody.DataAgreement.Policy
	toBeUpdatedDataAgreement.Purpose = requestBody.DataAgreement.Purpose
	toBeUpdatedDataAgreement.PurposeDescription = requestBody.DataAgreement.PurposeDescription
	toBeUpdatedDataAgreement.LawfulBasis = requestBody.DataAgreement.LawfulBasis
	toBeUpdatedDataAgreement.MethodOfUse = requestBody.DataAgreement.MethodOfUse
	toBeUpdatedDataAgreement.DpiaDate = requestBody.DataAgreement.DpiaDate
	toBeUpdatedDataAgreement.DpiaSummaryUrl = requestBody.DataAgreement.DpiaSummaryUrl
	toBeUpdatedDataAgreement.Signature = requestBody.DataAgreement.Signature
	toBeUpdatedDataAgreement.Active = requestBody.DataAgreement.Active
	toBeUpdatedDataAgreement.Forgettable = requestBody.DataAgreement.Forgettable
	toBeUpdatedDataAgreement.CompatibleWithVersionId = requestBody.DataAgreement.CompatibleWithVersionId
	toBeUpdatedDataAgreement.Lifecycle = requestBody.DataAgreement.Lifecycle

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

	// Repository
	daRepo := dataagreement.DataAgreementRepository{}
	daRepo.Init(organisationId)
	// Get policy from db
	toBeUpdatedDataAgreement, err := daRepo.Get(dataAgreementId)
	if err != nil {
		common.HandleErrorV2(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	// Get revision from db
	currentRevision, err := revision.GetLatestByDataAgreementId(dataAgreementId)
	if err != nil {
		common.HandleErrorV2(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	// Update data agreement from request body
	toBeUpdatedDataAgreement = updateDataAgreementFromRequestBody(dataAgreementReq, toBeUpdatedDataAgreement)

	// Bump major version for policy
	updatedVersion, err := common.BumpMajorVersion(toBeUpdatedDataAgreement.Version)
	if err != nil {
		m := fmt.Sprintf("Failed to bump major version for data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	toBeUpdatedDataAgreement.Version = updatedVersion

	// Update revision
	newRevision, err := revision.UpdateRevisionForDataAgreement(toBeUpdatedDataAgreement, &currentRevision, orgAdminId)
	if err != nil {
		m := fmt.Sprintf("Failed to update revision for data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Save the data agreement to db
	savedDataAgreement, err := daRepo.Update(toBeUpdatedDataAgreement)
	if err != nil {
		m := fmt.Sprintf("Failed to update data agreement: %v", toBeUpdatedDataAgreement.Id)
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

	// Save the revision to db
	savedRevision, err := revision.Add(newRevision)
	if err != nil {
		m := fmt.Sprintf("Failed to create new data agreement: %v", newRevision.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Constructing the response
	var resp updateDataAgreementResp
	resp.DataAgreement = savedDataAgreement

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(savedRevision)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
