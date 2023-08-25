package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/image"
	"github.com/bb-consent/api/src/notifications"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/orgtype"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/user"
	"github.com/globalsign/mgo/bson"
	"github.com/gorilla/mux"
)

type organization struct {
	Organization org.Organization
}

// Organization organization data type
type orgRequest struct {
	Name        string `valid:"required"`
	Location    string `valid:"required"`
	TypeID      string `valid:"required"`
	Description string
	EulaURL     string
	HlcSupport  bool
}

// AddOrganization Adds an organization
func AddOrganization(w http.ResponseWriter, r *http.Request) {
	var orgReq orgRequest
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &orgReq)

	// validating request payload
	valid, err := govalidator.ValidateStruct(orgReq)
	if !valid {
		log.Printf("Missing mandatory params for adding organization")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// checking if the string contained whitespace only
	if strings.TrimSpace(orgReq.Name) == "" {
		m := "Failed to add organization: Missing mandatory param - Name"
		common.HandleError(w, http.StatusBadRequest, m, errors.New("missing mandatory param - Name"))
		return
	}

	if strings.TrimSpace(orgReq.Location) == "" {
		m := "Failed to add organization: Missing mandatory param - Location"
		common.HandleError(w, http.StatusBadRequest, m, errors.New("missing mandatory param - Location"))
		return
	}

	if strings.TrimSpace(orgReq.TypeID) == "" {
		m := "Failed to add organization: Missing mandatory param - TypeID"
		common.HandleError(w, http.StatusBadRequest, m, errors.New("missing mandatory param - TypeID"))
		return
	}

	orgType, err := orgtype.Get(orgReq.TypeID)
	if err != nil {
		m := fmt.Sprintf("Invalid organization type ID: %v", orgReq.TypeID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	admin := org.Admin{UserID: token.GetUserID(r), RoleID: common.GetRoleID("Admin")}

	var o org.Organization
	o.Name = orgReq.Name
	o.Location = orgReq.Location
	o.Type = orgType
	o.Description = orgReq.Description
	o.EulaURL = orgReq.EulaURL
	o.Admins = append(o.Admins, admin)
	o.HlcSupport = orgReq.HlcSupport

	orgResp, err := org.Add(o)
	if err != nil {
		m := fmt.Sprintf("Failed to add organization: %v", orgReq.Name)
		common.HandleError(w, http.StatusConflict, m, err)
		return
	}

	//Update user role for this organization
	_, err = user.AddRole(token.GetUserID(r), user.Role{RoleID: common.GetRoleID("Admin"), OrgID: orgResp.ID.Hex()})
	if err != nil {
		m := fmt.Sprintf("Failed to update user : %v roles for org: %v", token.GetUserID(r), orgResp.ID.Hex())
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(organization{orgResp})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

// GetOrganizationByID Gets a single organization by given id
func GetOrganizationByID(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]
	o, err := org.Get(organizationID)

	if err != nil {
		m := fmt.Sprintf("Failed to get organization by ID :%v", organizationID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(organization{o})
	w.Write(response)
}

type orgUpdateReq struct {
	Name        string
	Location    string
	Description string
	PolicyURL   string
}

// UpdateOrganization Updates an organization
func UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	var orgUpReq orgUpdateReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &orgUpReq)

	organizationID := mux.Vars(r)["organizationID"]

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	if strings.TrimSpace(orgUpReq.Name) != "" {
		o.Name = orgUpReq.Name
	}
	if strings.TrimSpace(orgUpReq.Location) != "" {
		o.Location = orgUpReq.Location
	}
	if strings.TrimSpace(orgUpReq.Description) != "" {
		o.Description = orgUpReq.Description
	}
	if strings.TrimSpace(orgUpReq.PolicyURL) != "" {
		o.PolicyURL = orgUpReq.PolicyURL
	}

	orgResp, err := org.Update(o)
	if err != nil {
		m := fmt.Sprintf("Failed to update organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	go user.UpdateOrganizationsSubscribedUsers(orgResp)
	//response, _ := json.Marshal(organization{orgResp})
	//w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	//w.Write(response)
}

// UpdateOrganizationCoverImage Inserts the image and update the id to user
func UpdateOrganizationCoverImage(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]

	file, _, err := r.FormFile("orgimage")
	if err != nil {
		m := fmt.Sprintf("Failed to extract image organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	if err != nil {
		m := fmt.Sprintf("Failed to copy image organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	imageID, err := image.Add(buf.Bytes())
	if err != nil {
		m := fmt.Sprintf("Failed to store image in data store organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	imageURL := "https://" + r.Host + "/v1/organizations/" + organizationID + "/image/" + imageID
	o, err := org.UpdateCoverImage(organizationID, imageID, imageURL)
	if err != nil {
		m := fmt.Sprintf("Failed to update organization: %v with image: %v details", organizationID, imageID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(organization{o})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// UpdateOrganizationLogoImage Inserts the image and update the id to user
func UpdateOrganizationLogoImage(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]

	file, _, err := r.FormFile("orgimage")
	if err != nil {
		m := fmt.Sprintf("Failed to extract image organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, file)
	if err != nil {
		m := fmt.Sprintf("Failed to copy image organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	imageID, err := image.Add(buf.Bytes())
	if err != nil {
		m := fmt.Sprintf("Failed to store image in data store organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	imageURL := "https://" + r.Host + "/v1/organizations/" + organizationID + "/image/" + imageID
	o, err := org.UpdateLogoImage(organizationID, imageID, imageURL)
	if err != nil {
		m := fmt.Sprintf("Failed to update organization: %v with image: %v details", organizationID, imageID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(organization{o})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// GetOrganizationImage Retrieves the organization image
func GetOrganizationImage(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]
	imageID := mux.Vars(r)["imageID"]

	image, err := image.Get(imageID)

	if err != nil {
		m := fmt.Sprintf("Failed to fetch image with id: %v for org: %v", imageID, organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(image.Data)
}

type orgEulaUpReq struct {
	EulaURL string `valid:"required,url"`
}

// UpdateOrgEula Updates an organization EULA URL
func UpdateOrgEula(w http.ResponseWriter, r *http.Request) {
	var orgUpReq orgEulaUpReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &orgUpReq)

	// validating request params
	valid, err := govalidator.ValidateStruct(orgUpReq)

	if !valid {
		log.Printf("Missing mandatory param for updating EULA for org")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	organizationID := mux.Vars(r)["organizationID"]

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	o.EulaURL = orgUpReq.EulaURL

	orgResp, err := org.Update(o)
	if err != nil {
		m := fmt.Sprintf("Failed to update organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	go handleEulaUpdateNotification(orgResp)

	//response, _ := json.Marshal(organization{orgResp})
	//w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	//w.Write(response)
}

// DeleteOrgEula Updates an organization EULA URL
func DeleteOrgEula(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	o.EulaURL = ""

	orgResp, err := org.Update(o)
	if err != nil {
		m := fmt.Sprintf("Failed to update organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	go handleEulaUpdateNotification(orgResp)

	//response, _ := json.Marshal(organization{orgResp})
	//w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	//w.Write(response)
}

// TODO: Group these to err, info and introduce global counters
var consentGetErrCount = 0
var notificationErrCount = 0
var notificationSent = 0

func handleEulaUpdateNotification(o org.Organization) {
	// Get all users subscribed to this organization.
	orgID := o.ID.Hex()

	iter := user.GetOrgSubscribeIter(orgID)

	var u user.User

	for iter.Next(&u) {
		if u.Client.Token == "" {
			continue
		}
		err := notifications.SendEulaUpdateNotification(u, o)
		if err != nil {
			notificationErrCount++
			continue
		}
		notificationSent++
	}
	log.Printf("notification sending for EULA update orgID: %v with err: %v sent: %v", orgID,
		notificationErrCount, notificationSent)

	err := iter.Close()
	if err != nil {
		log.Printf("Failed to close the iterator: %v", iter)
	}
}

type adminReq struct {
	UserID string `valid:"required"`
	RoleID int    `valid:"required"`
}

// AddOrgAdmin Add admins, dpo and other roles to organization users
func AddOrgAdmin(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]

	var aReq adminReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &aReq)

	//TODO: Validate the struct
	valid, err := govalidator.ValidateStruct(aReq)
	if valid != true {
		log.Printf("Missing mandatory params for adding organization admin")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// validating UserID provided
	_, err = user.Get(aReq.UserID)
	if err != nil {
		m := fmt.Sprintf("Failed to add admin user to organization: %v invalid UserID: %v", organizationID, aReq.UserID)
		common.HandleError(w, http.StatusBadRequest, m, nil)
		return
	}

	if !common.IsValidRoleID(aReq.RoleID) {
		m := fmt.Sprintf("Failed to add admin user(%v) to organization: %v invalid RoleID: %v", aReq.UserID, organizationID, aReq.RoleID)
		common.HandleError(w, http.StatusBadRequest, m, nil)
		return
	}

	addAdminReq := org.Admin{UserID: aReq.UserID, RoleID: aReq.RoleID}
	o, err := org.AddAdminUsers(organizationID, addAdminReq)
	if err != nil {
		m := fmt.Sprintf("Failed to add admin user(%v) to organization: %v", aReq.UserID, organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	userOrg := user.Org{OrgID: o.ID, Name: o.Name, Location: o.Location, Type: o.Type.Type, TypeID: o.Type.ID}
	_, err = user.UpdateOrganization(aReq.UserID, userOrg)
	if err != nil {
		m := fmt.Sprintf("Failed to add user(%v) to organization: %v", aReq.UserID, organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	_, err = user.AddRole(aReq.UserID, user.Role{RoleID: aReq.RoleID, OrgID: organizationID})
	if err != nil {
		m := fmt.Sprintf("Failed to set user(%v) as %v to organization: %v", aReq.UserID, common.GetRole(aReq.RoleID), organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(organization{o})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// GetOrgAdmins Get all the special users admin/dpo/dev of this orgs
func GetOrgAdmins(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]

	o, err := org.GetAdminUsers(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to get admin users of organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	type orgAdmin struct {
		UserID string
		Role   string
	}

	var orgAdmins []orgAdmin
	for _, admin := range o.Admins {
		orgAdmins = append(orgAdmins, orgAdmin{admin.UserID, common.GetRole(admin.RoleID).Role})
	}
	response, _ := json.Marshal(orgAdmins)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// DeleteOrgAdmin Delete admins from organization
func DeleteOrgAdmin(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]

	var aReq adminReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &aReq)

	//TODO: Validate the struct
	valid, err := govalidator.ValidateStruct(aReq)
	if valid != true {
		log.Printf("Missing mandatory params for deleting organization admin")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	deleteAdminReq := org.Admin{RoleID: aReq.RoleID, UserID: aReq.UserID}
	o, err := org.DeleteAdminUsers(organizationID, deleteAdminReq)
	if err != nil {
		m := fmt.Sprintf("Failed to delete admin user(%v) from organization: %v", aReq, organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	_, err = user.DeleteOrganization(aReq.UserID, o.ID.Hex())
	if err != nil {
		m := fmt.Sprintf("Failed to delete admin user(%v) from organization: %v", aReq, organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	_, err = user.RemoveRole(aReq.UserID, user.Role{RoleID: aReq.RoleID, OrgID: o.ID.Hex()})

	response, _ := json.Marshal(organization{o})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// GetOrganizationRoles Get the list of organization roles
func GetOrganizationRoles(w http.ResponseWriter, r *http.Request) {
	roles := common.GetRoles()

	response, _ := json.Marshal(roles)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

type purpose struct {
	Name                    string `valid:"required"`
	Description             string `valid:"required"`
	LawfulBasisOfProcessing int
	PolicyURL               string `valid:"required"`
	AttributeType           int
	Jurisdiction            string
	Disclosure              string
	IndustryScope           string
	DataRetention           org.DataRetention
	Restriction             string
	Shared3PP               bool
	SSIID                   string
}

type purposeReq struct {
	Purposes []purpose
}

// AddConsentPurposes Adds consent purpose to the organization
func AddConsentPurposes(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var pReq purposeReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &pReq)

	// validating request payload
	valid, err := govalidator.ValidateStruct(pReq)
	if !valid {
		log.Printf("Missing mandatory fields for a adding consent purpose to org")
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	for _, p := range pReq.Purposes {

		// Proceed if lawful basis of processing provided is valid
		if !isValidLawfulBasisOfProcessing(p.LawfulBasisOfProcessing) {
			continue
		}

		tempLawfulUsage := getLawfulUsageByLawfulBasis(p.LawfulBasisOfProcessing)

		tempPurpose := org.Purpose{
			ID:                      bson.NewObjectId().Hex(),
			Name:                    p.Name,
			Description:             p.Description,
			LawfulUsage:             tempLawfulUsage,
			LawfulBasisOfProcessing: p.LawfulBasisOfProcessing,
			PolicyURL:               p.PolicyURL,
			AttributeType:           p.AttributeType,
			Jurisdiction:            p.Jurisdiction,
			Disclosure:              p.Disclosure,
			IndustryScope:           p.IndustryScope,
			DataRetention:           p.DataRetention,
			Restriction:             p.Restriction,
			Shared3PP:               p.Shared3PP,
			SSIID:                   p.SSIID}

		o.Purposes = append(o.Purposes, tempPurpose)
	}

	orgResp, err := org.UpdatePurposes(o.ID.Hex(), o.Purposes)
	if err != nil {
		m := fmt.Sprintf("Failed to update purpose to organization: %v", o.Name)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}
	/*
		u, err := user.Get(token.GetUserID(r))
		if err != nil {
			//notifications.SendPurposeUpdateNotification(u, o.ID.Hex(), )
		}
	*/

	response, _ := json.Marshal(organization{orgResp})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

type getPurposesResp struct {
	OrgID    string
	Purposes []org.Purpose
}

// GetPurposes Gets an organization purposes
func GetPurposes(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(getPurposesResp{OrgID: o.ID.Hex(), Purposes: o.Purposes})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

// DeleteConsentPurposeByID Deletes the given purpose by ID
func DeleteConsentPurposeByID(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]
	purposeID := mux.Vars(r)["purposeID"]

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var purposeToDelete org.Purpose
	for _, p := range o.Purposes {
		if p.ID == purposeID {
			purposeToDelete = p
		}
	}

	if purposeToDelete == (org.Purpose{}) {
		m := fmt.Sprintf("Failed to find purpose with ID: %v in organization: %v", purposeID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	//TODO: Before we delete purpose, need to remove the purpose from the templates
	err = deletePurposeIDFromTemplate(purposeID, o.ID.Hex(), o.Templates)
	if err != nil {
		m := fmt.Sprintf("Failed to update template for organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	orgResp, err := org.DeletePurposes(o.ID.Hex(), purposeToDelete)
	if err != nil {
		m := fmt.Sprintf("Failed to delete purpose: %v from organization: %v", purposeID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	response, _ := json.Marshal(organization{orgResp})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

// UpdatePurposeByID Update the given purpose by ID
func UpdatePurposeByID(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]
	purposeID := mux.Vars(r)["purposeID"]

	var uReq purpose
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &uReq)

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// Proceed if lawful basis of processing provided is valid
	if !isValidLawfulBasisOfProcessing(uReq.LawfulBasisOfProcessing) {
		m := fmt.Sprintf("Invalid lawful basis of processing provided")
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	tempLawfulUsage := getLawfulUsageByLawfulBasis(uReq.LawfulBasisOfProcessing)

	var found = false
	for i := range o.Purposes {
		if o.Purposes[i].ID == purposeID {
			found = true
			o.Purposes[i].Name = strings.TrimSpace(uReq.Name)
			o.Purposes[i].Description = strings.TrimSpace(uReq.Description)
			o.Purposes[i].PolicyURL = strings.TrimSpace(uReq.PolicyURL)
			o.Purposes[i].LawfulUsage = tempLawfulUsage
			o.Purposes[i].LawfulBasisOfProcessing = uReq.LawfulBasisOfProcessing
			o.Purposes[i].Jurisdiction = uReq.Jurisdiction
			o.Purposes[i].Disclosure = uReq.Disclosure
			o.Purposes[i].IndustryScope = uReq.IndustryScope
			o.Purposes[i].DataRetention = uReq.DataRetention
			o.Purposes[i].Restriction = uReq.Restriction
			o.Purposes[i].Shared3PP = uReq.Shared3PP
			if (o.Purposes[i].AttributeType != uReq.AttributeType) ||
				(o.Purposes[i].SSIID != uReq.SSIID) {
				m := fmt.Sprintf("Can not modify attributeType or SSIID for purpose: %v organization: %v",
					organizationID, purposeID)
				common.HandleError(w, http.StatusBadRequest, m, err)
				return
			}
		}
	}

	if !found {
		m := fmt.Sprintf("Failed to find purpose with ID: %v in organization: %v", purposeID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	o, err = org.Update(o)
	if err != nil {
		m := fmt.Sprintf("Failed to update purpose: %v in organization: %v", purposeID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	response, _ := json.Marshal(organization{o})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

// GetPurposeByID Get a purpose by ID
func GetPurposeByID(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["organizationID"]
	purposeID := mux.Vars(r)["purposeID"]

	o, err := org.Get(orgID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", orgID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	type purposeTemplates struct {
		ID      string
		Consent string
	}

	type purposeDetails struct {
		Purpose   org.Purpose
		Templates []purposeTemplates
	}
	var pDetails purposeDetails
	for _, p := range o.Purposes {
		if p.ID == purposeID {
			pDetails.Purpose = p
		}
	}

	for _, t := range o.Templates {
		for _, pID := range t.PurposeIDs {
			if pID == purposeID {
				pDetails.Templates = append(pDetails.Templates, purposeTemplates{ID: t.ID, Consent: t.Consent})
			}
		}
	}

	response, _ := json.Marshal(pDetails)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// Check if the lawful usage ID provided is valid
func isValidLawfulBasisOfProcessing(lawfulBasis int) bool {
	isFound := false
	for _, lawfulBasisOfProcessingMapping := range org.LawfulBasisOfProcessingMappings {
		if lawfulBasisOfProcessingMapping.ID == lawfulBasis {
			isFound = true
			break
		}
	}

	return isFound
}

// Fetch the lawful usage based on the lawful basis ID
func getLawfulUsageByLawfulBasis(lawfulBasis int) bool {
	if lawfulBasis == org.ConsentBasis {
		return false
	} else {
		return true
	}
}

func deletePurposeIDFromTemplate(purposeID string, orgID string, templates []org.Template) error {
	for _, t := range templates {
		for _, p := range t.PurposeIDs {
			if p == purposeID {
				var template org.Template
				template.Consent = t.Consent
				template.ID = t.ID
				for _, p := range t.PurposeIDs {
					if p != purposeID {
						template.PurposeIDs = append(template.PurposeIDs, p)
					}
				}
				_, err := org.DeleteTemplates(orgID, t)
				if err != nil {
					fmt.Printf("Failed to delete template: %v from organization: %v", t.ID, orgID)
					return err
				}
				if len(template.PurposeIDs) == 0 {
					continue
				}
				err = org.AddTemplates(orgID, template)
				if err != nil {
					fmt.Printf("Failed to add template: %v from organization: %v", t.ID, orgID)
					return err
				}
				continue
			}
		}
	}
	return nil
}

type template struct {
	Consent    string   `valid:"required"`
	PurposeIDs []string `valid:"required"`
}
type templateReq struct {
	Templates []template
}

// AddConsentTemplates Adds an organization template
func AddConsentTemplates(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var tReq templateReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &tReq)

	// validating request payload
	valid, err := govalidator.ValidateStruct(tReq)
	if !valid {
		log.Printf("Missing mandatory fields for adding consent template to org: %v", organizationID)
		common.HandleError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	// validating purposeIDs provided
	for _, t := range tReq.Templates {
		// checking if purposeID provided exist in the org
		for _, p := range t.PurposeIDs {
			_, err = org.GetPurpose(organizationID, p)
			if err != nil {
				m := fmt.Sprintf("Invalid purposeID:%v provided;Failed to update templates to organization: %v", p, o.Name)
				common.HandleError(w, http.StatusBadRequest, m, err)
				return
			}
		}

		// Appending the new template to existing org templates
		o.Templates = append(o.Templates, org.Template{
			ID:         bson.NewObjectId().Hex(),
			Consent:    t.Consent,
			PurposeIDs: t.PurposeIDs,
		})
	}

	orgResp, err := org.UpdateTemplates(o.ID.Hex(), o.Templates)
	if err != nil {
		m := fmt.Sprintf("Failed to update templates to organization: %v", o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	response, _ := json.Marshal(organization{orgResp})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

type getTemplatesResp struct {
	OrgID     string
	Templates []org.Template
}

// GetTemplates Gets an organization templates
func GetTemplates(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(getTemplatesResp{OrgID: o.ID.Hex(), Templates: o.Templates})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

// DeleteConsentTemplateByID Deletes an organization templates
func DeleteConsentTemplateByID(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]
	templateID := mux.Vars(r)["templateID"]

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var templateToDelete org.Template
	for _, t := range o.Templates {
		if t.ID == templateID {
			templateToDelete = t
		}
	}

	if templateToDelete.ID != templateID {
		m := fmt.Sprintf("Failed to find template with ID: %v in organization: %v", templateID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	orgResp, err := org.DeleteTemplates(o.ID.Hex(), templateToDelete)
	if err != nil {
		m := fmt.Sprintf("Failed to delete template: %v from organization: %v", templateID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	response, _ := json.Marshal(organization{orgResp})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

// DeleteConsentTemplatesByID Deletes an organization templates
func DeleteConsentTemplatesByID(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]
	purposeID := mux.Vars(r)["purposeID"]

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// Validating provided purposeID
	isPurposeFound := false
	for _, p := range o.Purposes {
		if p.ID == purposeID {
			isPurposeFound = true
		}
	}

	if !isPurposeFound {
		m := fmt.Sprintf("Failed to find purpose:%v for organization: %v", purposeID, organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var toBeDeletedTemplateIDs []string
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &toBeDeletedTemplateIDs)

	toBeDeletedTemplateIndexs := make([]int, 0)

	for _, toBeDeletedTemplateID := range toBeDeletedTemplateIDs {
		for templateIndex, orgTemplate := range o.Templates {

			if orgTemplate.ID == toBeDeletedTemplateID {

				toBeDeletedTemplateIndexs = append(toBeDeletedTemplateIndexs, templateIndex)

			}
		}

	}

	toBeNotDeletedTemplates := make([]org.Template, 0)
	for _, templateIndex := range toBeDeletedTemplateIndexs {
		templatePurposeIDs := o.Templates[templateIndex].PurposeIDs

		for templatePurposeIDIndex, templatePurposeID := range templatePurposeIDs {
			if templatePurposeID == purposeID {
				o.Templates[templateIndex].PurposeIDs = removeElementFromSlice(o.Templates[templateIndex].PurposeIDs, templatePurposeIDIndex)
			}

		}

	}

	for _, tempOrgTemplate := range o.Templates {
		if len(tempOrgTemplate.PurposeIDs) > 0 {
			toBeNotDeletedTemplates = append(toBeNotDeletedTemplates, tempOrgTemplate)
		}
	}

	o.Templates = toBeNotDeletedTemplates
	o, err = org.Update(o)
	if err != nil {
		m := fmt.Sprintf("Failed to delete templates in organization: %v", o.ID.Hex())
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	if o.HlcSupport {
		log.Println("Haven't implemented ledger integrations")
	}

	response, _ := json.Marshal(organization{o})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

type getTemplateResp struct {
	OrgID    string
	Template org.Template
}

// GetTemplateByID Gets an organization templates
func GetTemplateByID(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]
	templateID := mux.Vars(r)["templateID"]

	t, err := org.GetTemplate(organizationID, templateID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v tempalte: %v", organizationID, templateID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(getTemplateResp{OrgID: organizationID, Template: t})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

// UpdateTemplateByID Updates an organization consent templates
func UpdateTemplateByID(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]
	templateID := mux.Vars(r)["templateID"]

	var uReq template
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &uReq)

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// validating PurposeIDs provided
	if uReq.PurposeIDs != nil {

		for _, p := range uReq.PurposeIDs {
			_, err := org.GetPurpose(organizationID, p)
			if err != nil {
				m := fmt.Sprintf("Invalid purposeID:%v provided;Failed to update template to organization: %v", p, o.Name)
				common.HandleError(w, http.StatusBadRequest, m, err)
				return
			}
		}
	}

	var found = false

	for i := range o.Templates {
		if o.Templates[i].ID == templateID {
			found = true

			// for partial updation
			if strings.TrimSpace(uReq.Consent) != "" {
				o.Templates[i].Consent = uReq.Consent
			}
			// only updating if any valid purpose id was given
			if uReq.PurposeIDs != nil {
				o.Templates[i].PurposeIDs = uReq.PurposeIDs
			}
		}
	}

	if !found {
		m := fmt.Sprintf("Failed to find template with ID: %v in organization: %v", templateID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	o, err = org.Update(o)
	if err != nil {
		m := fmt.Sprintf("Failed to update template: %v in organization: %v", templateID, o.Name)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	response, _ := json.Marshal(organization{o})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

func removeElementFromSlice(slice []string, index int) []string {

	newSlice := make([]string, 0)

	for tempIndex, _ := range slice {
		if tempIndex != index {
			newSlice = append(newSlice, slice[tempIndex])
		}
	}

	return newSlice
}

type globalPolicyConfigurationResp struct {
	PolicyURL     string
	DataRetention org.DataRetention
	Jurisdiction  string
	Disclosure    string
	Type          orgtype.OrgType
	Restriction   string
	Shared3PP     bool
}

// GetGlobalPolicyConfiguration Handler to get global policy configurations
func GetGlobalPolicyConfiguration(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// Constructing the response
	var resp globalPolicyConfigurationResp

	resp.PolicyURL = o.PolicyURL
	resp.DataRetention = o.DataRetention

	if len(strings.TrimSpace(o.Jurisdiction)) == 0 {
		resp.Jurisdiction = o.Location
		o.Jurisdiction = o.Location
	} else {
		resp.Jurisdiction = o.Jurisdiction
	}

	if len(strings.TrimSpace(o.Disclosure)) == 0 {
		resp.Disclosure = "false"
		o.Disclosure = "false"
	} else {
		resp.Disclosure = o.Disclosure
	}

	resp.Type = o.Type
	resp.Restriction = o.Restriction
	resp.Shared3PP = o.Shared3PP

	// Updating global configuration policy with defaults
	_, err = org.Update(o)
	if err != nil {
		m := fmt.Sprintf("Failed to update global configuration with defaults to organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	response, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}

type globalPolicyConfigurationReq struct {
	PolicyURL       string
	RetentionPeriod int
	Jurisdiction    string
	Disclosure      string
	TypeID          string `valid:"required"`
	Restriction     string
	Shared3PP       bool
}

// UpdateGlobalPolicyConfiguration Handler to update global policy configuration
func UpdateGlobalPolicyConfiguration(w http.ResponseWriter, r *http.Request) {
	organizationID := mux.Vars(r)["organizationID"]

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var policyReq globalPolicyConfigurationReq
	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	json.Unmarshal(b, &policyReq)

	// Update global policy configuration for the org
	o.PolicyURL = policyReq.PolicyURL

	if len(strings.TrimSpace(policyReq.Jurisdiction)) != 0 {
		o.Jurisdiction = policyReq.Jurisdiction
	}

	o.Restriction = policyReq.Restriction
	o.Shared3PP = policyReq.Shared3PP

	if policyReq.Disclosure == "false" || policyReq.Disclosure == "true" {
		o.Disclosure = policyReq.Disclosure
	}

	// Check if type id is valid bson objectid hex
	if !bson.IsObjectIdHex(policyReq.TypeID) {
		m := fmt.Sprintf("Invalid organization type ID: %v", policyReq.TypeID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	orgType, err := orgtype.Get(policyReq.TypeID)
	if err != nil {
		m := fmt.Sprintf("Invalid organization type ID: %v", policyReq.TypeID)
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	o.Type = orgType

	if policyReq.RetentionPeriod > 0 {
		o.DataRetention.RetentionPeriod = int64(policyReq.RetentionPeriod)
		o.DataRetention.Enabled = true
	} else {
		o.DataRetention.RetentionPeriod = 0
		o.DataRetention.Enabled = false
	}

	// Updating global configuration policy with defaults
	o, err = org.Update(o)
	if err != nil {
		m := fmt.Sprintf("Failed to update global configuration to organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	// Constructing the response
	var resp globalPolicyConfigurationResp
	resp.PolicyURL = o.PolicyURL
	resp.DataRetention = o.DataRetention
	resp.Jurisdiction = o.Jurisdiction
	resp.Disclosure = o.Disclosure
	resp.Type = o.Type
	resp.Restriction = o.Restriction
	resp.Shared3PP = o.Shared3PP

	response, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
