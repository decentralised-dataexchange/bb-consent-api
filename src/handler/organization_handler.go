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
