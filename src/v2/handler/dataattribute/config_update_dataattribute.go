package dataattribute

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/v2/dataagreement"
	"github.com/bb-consent/api/src/v2/dataattribute"
	"github.com/bb-consent/api/src/v2/revision"
	"github.com/gorilla/mux"
)

func validateUpdateDataAttributeRequestBody(dataAttributeReq updateDataAttributeReq, organisationId string) error {
	// validating request payload
	valid, err := govalidator.ValidateStruct(dataAttributeReq)
	if !valid {
		return err
	}

	// Repository
	darepo := dataagreement.DataAgreementRepository{}
	darepo.Init(organisationId)

	// validating data agreement Ids provided
	// checking if data agreement Ids provided exist in the db
	for _, p := range dataAttributeReq.DataAttribute.AgreementIds {
		exists, err := darepo.IsDataAgreementExist(p)
		if err != nil || exists < 1 {
			m := fmt.Sprintf("Invalid data agreementId: %v provided;Failed to add data attribute", p)
			return errors.New(m)
		}
	}

	return nil
}

func updateDataAttributeFromRequestBody(requestBody updateDataAttributeReq, toBeUpdatedDataAttribute dataattribute.DataAttribute) dataattribute.DataAttribute {

	toBeUpdatedDataAttribute.AgreementIds = requestBody.DataAttribute.AgreementIds
	toBeUpdatedDataAttribute.Name = requestBody.DataAttribute.Name
	toBeUpdatedDataAttribute.Description = requestBody.DataAttribute.Description
	toBeUpdatedDataAttribute.Sensitivity = requestBody.DataAttribute.Sensitivity
	toBeUpdatedDataAttribute.Category = requestBody.DataAttribute.Category

	return toBeUpdatedDataAttribute
}

type updateDataAttributeReq struct {
	DataAttribute dataattribute.DataAttribute `json:"dataAttribute" valid:"required"`
}

type updateDataAttributeResp struct {
	DataAttribute dataattribute.DataAttribute `json:"dataAttribute"`
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
	dataAttributeRepo := dataattribute.DataAttributeRepository{}
	dataAttributeRepo.Init(organisationId)
	// Get data attribute from db
	toBeUpdatedDataAttribute, err := dataAttributeRepo.Get(dataAttributeId)
	if err != nil {
		common.HandleErrorV2(w, http.StatusInternalServerError, err.Error(), err)
		return
	}
	// Get data attribute from db
	currentRevision, err := revision.GetLatestByDataAttributeId(dataAttributeId)
	if err != nil {
		common.HandleErrorV2(w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	// Update data attribute from request body
	toBeUpdatedDataAttribute = updateDataAttributeFromRequestBody(dataAttributeReq, toBeUpdatedDataAttribute)

	// Bump major version for data attribute
	updatedVersion, err := common.BumpMajorVersion(toBeUpdatedDataAttribute.Version)
	if err != nil {
		m := fmt.Sprintf("Failed to bump major version for data attribute: %v", dataAttributeId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}
	toBeUpdatedDataAttribute.Version = updatedVersion

	// Update revision
	newRevision, err := revision.UpdateRevisionForDataAttribute(toBeUpdatedDataAttribute, &currentRevision, orgAdminId)
	if err != nil {
		m := fmt.Sprintf("Failed to update revision for data attribute: %v", dataAttributeId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Save the data attribute to db
	savedDataAttribute, err := dataAttributeRepo.Update(toBeUpdatedDataAttribute)
	if err != nil {
		m := fmt.Sprintf("Failed to update data attribute: %v", toBeUpdatedDataAttribute.Id)
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
		m := fmt.Sprintf("Failed to create new revision: %v", newRevision.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Constructing the response
	var resp updateDataAttributeResp
	resp.DataAttribute = savedDataAttribute

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(savedRevision)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
