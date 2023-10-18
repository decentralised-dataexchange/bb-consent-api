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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func validateAddDataAttributeRequestBody(dataAttributeReq addDataAttributeReq, organisationId string) error {
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

func updateDataAttributeFromAddDataAttributeRequestBody(requestBody addDataAttributeReq, newDataAttribute dataattribute.DataAttribute) dataattribute.DataAttribute {

	newDataAttribute.AgreementIds = requestBody.DataAttribute.AgreementIds
	newDataAttribute.Name = requestBody.DataAttribute.Name
	newDataAttribute.Description = requestBody.DataAttribute.Description
	newDataAttribute.Sensitivity = requestBody.DataAttribute.Sensitivity
	newDataAttribute.Category = requestBody.DataAttribute.Category

	return newDataAttribute
}

type addDataAttributeReq struct {
	DataAttribute dataattribute.DataAttribute `json:"dataAttribute" valid:"required"`
}

type addDataAttributeResp struct {
	DataAttribute dataattribute.DataAttribute `json:"dataAttribute"`
	Revision      interface{}                 `json:"revision"`
}

// ConfigCreateDataAttribute
func ConfigCreateDataAttribute(w http.ResponseWriter, r *http.Request) {
	// Current user
	orgAdminId := token.GetUserID(r)

	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	// Request body
	var dataAttributeReq addDataAttributeReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &dataAttributeReq)

	// Validate request body
	err := validateAddDataAttributeRequestBody(dataAttributeReq, organisationId)
	if err != nil {
		common.HandleErrorV2(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	version := common.IntegerToSemver(1)

	// Initialise data attribute
	var newDataAttribute dataattribute.DataAttribute
	newDataAttribute.Id = primitive.NewObjectID().Hex()
	// Update data attribute from request body
	newDataAttribute = updateDataAttributeFromAddDataAttributeRequestBody(dataAttributeReq, newDataAttribute)
	newDataAttribute.OrganisationId = organisationId
	newDataAttribute.IsDeleted = false
	newDataAttribute.Version = version

	// Create new revision
	newRevision, err := revision.CreateRevisionForDataAttribute(newDataAttribute, orgAdminId)
	if err != nil {
		m := fmt.Sprintf("Failed to create revision for new data attribute: %v", newDataAttribute.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Repository
	dataAttributeRepo := dataattribute.DataAttributeRepository{}
	dataAttributeRepo.Init(organisationId)

	// Save the data attribute to db
	savedDataAttribute, err := dataAttributeRepo.Add(newDataAttribute)
	if err != nil {
		m := fmt.Sprintf("Failed to create new data attribute: %v", newDataAttribute.Name)
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
	var resp addDataAttributeResp
	resp.DataAttribute = savedDataAttribute

	var revisionForHTTPResponse revision.RevisionForHTTPResponse
	revisionForHTTPResponse.Init(savedRevision)
	resp.Revision = revisionForHTTPResponse

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
