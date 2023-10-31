package dataattribute

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/token"
	"github.com/gorilla/mux"
)

func validateUpdateDataAttributeRequestBody(dataAttributeReq updateDataAttributeReq, organisationId string) error {
	// validating request payload
	valid, err := govalidator.ValidateStruct(dataAttributeReq)
	if !valid {
		return err
	}

	return nil
}

func updateDataAttributeFromRequestBody(dataAttributeId string, requestBody updateDataAttributeReq, toBeUpdatedDataAgreement dataagreement.DataAgreement) dataagreement.DataAgreement {

	var dataAttributes []dataagreement.DataAttribute

	for i := range toBeUpdatedDataAgreement.DataAttributes {
		var dataAttribute dataagreement.DataAttribute
		if toBeUpdatedDataAgreement.DataAttributes[i].Id.Hex() == dataAttributeId {
			dataAttribute.Id = toBeUpdatedDataAgreement.DataAttributes[i].Id
			dataAttribute.Name = requestBody.DataAttribute.Name
			dataAttribute.Description = requestBody.DataAttribute.Description
			dataAttribute.Sensitivity = requestBody.DataAttribute.Sensitivity
			dataAttribute.Category = requestBody.DataAttribute.Category
		} else {
			dataAttribute.Id = toBeUpdatedDataAgreement.DataAttributes[i].Id
			dataAttribute.Name = toBeUpdatedDataAgreement.DataAttributes[i].Name
			dataAttribute.Description = toBeUpdatedDataAgreement.DataAttributes[i].Description
			dataAttribute.Sensitivity = toBeUpdatedDataAgreement.DataAttributes[i].Sensitivity
			dataAttribute.Category = toBeUpdatedDataAgreement.DataAttributes[i].Category
		}
		dataAttributes = append(dataAttributes, dataAttribute)
	}
	toBeUpdatedDataAgreement.DataAttributes = dataAttributes

	return toBeUpdatedDataAgreement
}

func dataAttributeResp(dataAttributeId string, savedDataAgreement dataagreement.DataAgreement) dataagreement.DataAttribute {

	var dataAttribute dataagreement.DataAttribute

	for i := range savedDataAgreement.DataAttributes {

		if savedDataAgreement.DataAttributes[i].Id.Hex() == dataAttributeId {
			dataAttribute.Id = savedDataAgreement.DataAttributes[i].Id
			dataAttribute.Name = savedDataAgreement.DataAttributes[i].Name
			dataAttribute.Description = savedDataAgreement.DataAttributes[i].Description
			dataAttribute.Sensitivity = savedDataAgreement.DataAttributes[i].Sensitivity
			dataAttribute.Category = savedDataAgreement.DataAttributes[i].Category
			return dataAttribute
		}

	}

	return dataAttribute
}

type updateDataAttributeReq struct {
	DataAttribute dataagreement.DataAttribute `json:"dataAttribute" valid:"required"`
}

type updateDataAttributeResp struct {
	DataAttribute dataagreement.DataAttribute `json:"dataAttribute"`
	Revision      interface{}                 `json:"revision"`
}

// ConfigUpdateDataAttribute
func ConfigUpdateDataAttribute(w http.ResponseWriter, r *http.Request) {
	// Current user
	orgAdminId := token.GetUserID(r)

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Path params
	dataAttributeId := mux.Vars(r)[config.DataAttributeId]
	dataAttributeId = common.Sanitize(dataAttributeId)

	// Request body
	var dataAttributeReq updateDataAttributeReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &dataAttributeReq)

	// Validate request body
	err := validateUpdateDataAttributeRequestBody(dataAttributeReq, organisationId)
	if err != nil {
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Repository
	darepo := dataagreement.DataAgreementRepository{}
	darepo.Init(organisationId)

	// Get data attribute from db
	toBeUpdatedDataAgreement, err := darepo.GetByDataAttributeId(dataAttributeId)
	if err != nil {
		common.HandleErrorV2(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	var currentRevision revision.Revision
	// Get current revision from db
	currentRevision, err = revision.GetLatestByDataAgreementId(toBeUpdatedDataAgreement.Id.Hex())
	if err != nil {
		common.HandleErrorV2(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	// Update data attribute from request body
	toBeUpdatedDataAgreement = updateDataAttributeFromRequestBody(dataAttributeId, dataAttributeReq, toBeUpdatedDataAgreement)

	// Bump major version for data attribute
	updatedVersion, err := common.BumpMajorVersion(toBeUpdatedDataAgreement.Version)
	if err != nil {
		m := fmt.Sprintf("Failed to bump major version for data attribute: %v", dataAttributeId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	if toBeUpdatedDataAgreement.Active {

		toBeUpdatedDataAgreement.Version = updatedVersion

		// Update revision
		newRevision, err := revision.UpdateRevisionForDataAgreement(toBeUpdatedDataAgreement, &currentRevision, orgAdminId)
		if err != nil {
			m := fmt.Sprintf("Failed to update revision for data attribute: %v", dataAttributeId)
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
		currentRevision, err = revision.Add(newRevision)
		if err != nil {
			m := fmt.Sprintf("Failed to create new revision: %v", newRevision.Id)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
	}

	// Save the data agreement to db
	savedDataAgreement, err := darepo.Update(toBeUpdatedDataAgreement)
	if err != nil {
		m := fmt.Sprintf("Failed to update data attribute: %v", dataAttributeId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Constructing the response
	var resp updateDataAttributeResp
	updatedDataAttribute := dataAttributeResp(dataAttributeId, savedDataAgreement)

	resp.DataAttribute = updatedDataAttribute
	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(currentRevision)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
