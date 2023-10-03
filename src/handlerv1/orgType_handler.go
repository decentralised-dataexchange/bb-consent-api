package handlerv1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/image"
	"github.com/bb-consent/api/src/org"
	ot "github.com/bb-consent/api/src/orgtype"
	"github.com/bb-consent/api/src/user"
	"github.com/gorilla/mux"
)

type addOrgTypeReq struct {
	Type string `valid:"required"`
}

// AddOrganizationType Adds an organization type
func AddOrganizationType(w http.ResponseWriter, r *http.Request) {
	var addReq addOrgTypeReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &addReq)

	// validating request payload
	valid, err := govalidator.ValidateStruct(addReq)
	if valid != true {
		log.Printf("Missing mandatory params for adding organization")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	var o ot.OrgType
	o.Type = addReq.Type

	o, err = ot.Add(o)
	if err != nil {
		m := fmt.Sprintf("Failed to add organization type: %v", o.Type)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(o)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

type updateOrgTypeReq struct {
	Type string `valid:"required"`
}

// UpdateOrganizationType Updates an organization type
func UpdateOrganizationType(w http.ResponseWriter, r *http.Request) {
	organizationTypeID := mux.Vars(r)["typeID"]

	var updateReq updateOrgTypeReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &updateReq)

	// validating request payload
	valid, err := govalidator.ValidateStruct(updateReq)
	if valid != true {
		log.Printf("Missing mandatory params for updating organization type")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	orgType, err := ot.Update(organizationTypeID, updateReq.Type)
	if err != nil {
		m := fmt.Sprintf("Failed to update organization type: %v", organizationTypeID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	go org.UpdateOrganizationsOrgType(orgType)

	go user.UpdateOrgTypeOfSubscribedUsers(orgType)

	response, _ := json.Marshal(orgType)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

// DeleteOrganizationType Gets organization Type by given id
func DeleteOrganizationType(w http.ResponseWriter, r *http.Request) {
	typeID := mux.Vars(r)["typeID"]

	//TODO: Find all organizations with this type and then reject deletion if atleast one org exist.
	err := ot.Delete(typeID)
	if err != nil {
		m := fmt.Sprintf("Failed to delete organization type: %v", typeID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetOrganizationTypeByID Gets organization Type by given id
func GetOrganizationTypeByID(w http.ResponseWriter, r *http.Request) {
	typeID := mux.Vars(r)["typeID"]
	orgType, err := ot.Get(typeID)

	if err != nil {
		m := fmt.Sprintf("Failed to get organization type: %v", typeID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	response, _ := json.Marshal(orgType)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}

// GetOrganizationTypes Gets all organization types
func GetOrganizationTypes(w http.ResponseWriter, r *http.Request) {
	results, err := ot.GetAll()
	if err != nil {
		m := fmt.Sprintf("Failed to get organization types")
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(results)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.Write(response)
}

// UpdateOrganizationTypeImage Inserts the image and update the id to user
func UpdateOrganizationTypeImage(w http.ResponseWriter, r *http.Request) {
	organizationTypeID := mux.Vars(r)["typeID"]

	file, _, err := r.FormFile("orgtypeicon")
	if err != nil {
		m := fmt.Sprintf("Failed to extract image organizationType: %v", organizationTypeID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	if err != nil {
		m := fmt.Sprintf("Failed to copy image organizationType: %v", organizationTypeID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	imageID, err := image.Add(buf.Bytes())
	if err != nil {
		m := fmt.Sprintf("Failed to store image in data store organizationType: %v", organizationTypeID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	imageURL := "https://" + r.Host + "/v1/organizations/types/" + organizationTypeID + "/image"
	err = ot.UpdateImage(organizationTypeID, imageID, imageURL)
	if err != nil {
		m := fmt.Sprintf("Failed to update organizationType: %v with image: %v details", organizationTypeID, imageID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	orgType, err := ot.Get(organizationTypeID)
	if err != nil {
		m := fmt.Sprintf("Failed to update organizationType: %v with image: %v details", organizationTypeID, imageID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	go org.UpdateOrganizationsOrgType(orgType)
}

// GetOrganizationTypeImage Retrieves the organizationType image
func GetOrganizationTypeImage(w http.ResponseWriter, r *http.Request) {
	organizationTypeID := mux.Vars(r)["typeID"]

	//TODO: Get only the imageID from the organizationType and not the whole organizationType.
	organizationType, err := ot.Get(organizationTypeID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organizationType: %v", organizationTypeID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	image, err := image.Get(organizationType.ImageID)

	if err != nil {
		m := fmt.Sprintf("Failed to fetch image with id: %v", organizationType.ImageID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	w.Header().Set(config.ContentTypeHeader, config.ContentTypeImage)
	w.Write(image.Data)
}
