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

func validateReq(dataAttributeReq updateDataAttributeReq, organisationId string) error {
	// validating request payload
	valid, err := govalidator.ValidateStruct(dataAttributeReq)
	if !valid {
		return err
	}

	return nil
}

func updateDataAttributeFromReq(dataAttributeId string, requestBody updateDataAttributeReq, dataAttributes []dataagreement.DataAttribute) ([]dataagreement.DataAttribute, int) {
	updatedDataAttributes := make([]dataagreement.DataAttribute, len(dataAttributes))
	updatedDataAttributes = dataAttributes

	for i, dataAttribute := range dataAttributes {
		if dataAttribute.Id == dataAttributeId {
			updatedDataAttributes[i].Name = requestBody.DataAttribute.Name
			updatedDataAttributes[i].Description = requestBody.DataAttribute.Description
			updatedDataAttributes[i].Sensitivity = requestBody.DataAttribute.Sensitivity
			updatedDataAttributes[i].Category = requestBody.DataAttribute.Category

			return updatedDataAttributes, i
		}
	}
	return updatedDataAttributes, -1
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
	err := validateReq(dataAttributeReq, organisationId)
	if err != nil {
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// Repository
	darepo := dataagreement.DataAgreementRepository{}
	darepo.Init(organisationId)

	// Query data agreement by data attribute id
	toBeUpdatedDataAgreement, err := darepo.GetByDataAttributeId(dataAttributeId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement by data attribute id: %v", dataAttributeId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	currentVersion := toBeUpdatedDataAgreement.Version
	currentActiveStatus := toBeUpdatedDataAgreement.Active

	// Set data attribute from request body
	updatedDataAttributes, matchedIndex := updateDataAttributeFromReq(dataAttributeId, dataAttributeReq, toBeUpdatedDataAgreement.DataAttributes)
	toBeUpdatedDataAgreement.DataAttributes = updatedDataAttributes

	// Bump major version for data agreement
	updatedVersion, err := common.BumpMajorVersion(toBeUpdatedDataAgreement.Version)
	if err != nil {
		m := fmt.Sprintf("Failed to bump major version for data agreement: %v", toBeUpdatedDataAgreement.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Increment data agreement version
	toBeUpdatedDataAgreement.Version = updatedVersion

	// if data agreement is in draft mode then
	// version is not incremented
	if !currentActiveStatus {
		toBeUpdatedDataAgreement.Version = currentVersion
	}

	// Revision handling for data agreements
	// If data agreement is published then:
	// a. Add a new revision
	if toBeUpdatedDataAgreement.Active {

		// Update revision
		_, err = revision.UpdateRevisionForDataAgreement(toBeUpdatedDataAgreement, orgAdminId)
		if err != nil {
			m := fmt.Sprintf("Failed to update data agreement: %v", toBeUpdatedDataAgreement.Id)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

	}

	// Save the data agreement to db
	_, err = darepo.Update(toBeUpdatedDataAgreement)
	if err != nil {
		m := fmt.Sprintf("Failed to update data attribute: %v", dataAttributeId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Constructing the response
	var resp updateDataAttributeResp
	resp.DataAttribute = updatedDataAttributes[matchedIndex]
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
