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

func updateDataAttributeFromRequestBody(dataAttributeId string, requestBody updateDataAttributeReq, dataAttributes []dataagreement.DataAttribute) []dataagreement.DataAttribute {
	updatedDataAttributes := make([]dataagreement.DataAttribute, len(dataAttributes))

	for i, dataAttribute := range dataAttributes {
		updatedDataAttribute := dataAttribute
		if dataAttribute.Id.Hex() == dataAttributeId {
			updatedDataAttribute.Name = requestBody.DataAttribute.Name
			updatedDataAttribute.Description = requestBody.DataAttribute.Description
			updatedDataAttribute.Sensitivity = requestBody.DataAttribute.Sensitivity
			updatedDataAttribute.Category = requestBody.DataAttribute.Category
		}
		updatedDataAttributes[i] = updatedDataAttribute
	}
	return updatedDataAttributes
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

	// Get data agreement by data attribute id
	toBeUpdatedDataAgreement, err := darepo.GetByDataAttributeId(dataAttributeId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement by data attribute id: %v", dataAttributeId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Update data attribute if a match is found
	toBeUpdatedDataAgreement.DataAttributes = updateDataAttributeFromRequestBody(dataAttributeId, dataAttributeReq, toBeUpdatedDataAgreement.DataAttributes)

	// Revision handling for data agreements
	// If data agreement is published then:
	// a. Update data agreement version
	// b. Add a new revision
	if toBeUpdatedDataAgreement.Active {
		// Bump major version for data agreement
		updatedVersion, err := common.BumpMajorVersion(toBeUpdatedDataAgreement.Version)
		if err != nil {
			m := fmt.Sprintf("Failed to bump major version for data agreement: %v", toBeUpdatedDataAgreement.Id.Hex())
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

		// Increment data agreement version
		toBeUpdatedDataAgreement.Version = updatedVersion

		// Update revision
		_, err = revision.UpdateRevisionForDataAgreement(toBeUpdatedDataAgreement, orgAdminId)
		if err != nil {
			m := fmt.Sprintf("Failed to update data agreement: %v", toBeUpdatedDataAgreement.Id.Hex())
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
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
