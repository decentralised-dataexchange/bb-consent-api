package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/org"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

type addPurposeReq struct {
	purpose
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

// AddDataAgreement Adds a single data agreement to the organization
func AddDataAgreement(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Header.Get(config.OrganizationId)

	o, err := org.Get(organizationID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization: %v", organizationID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	var pReq addPurposeReq
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

	// Proceed if lawful basis of processing provided is valid
	if !isValidLawfulBasisOfProcessing(pReq.LawfulBasisOfProcessing) {
		m := fmt.Sprintf("Invalid lawful basis of processing provided")
		common.HandleError(w, http.StatusBadRequest, m, err)
		return
	}

	tempLawfulUsage := getLawfulUsageByLawfulBasis(pReq.LawfulBasisOfProcessing)

	newPurpose := org.Purpose{
		ID:                      primitive.NewObjectID().Hex(),
		Name:                    pReq.Name,
		Description:             pReq.Description,
		LawfulUsage:             tempLawfulUsage,
		LawfulBasisOfProcessing: pReq.LawfulBasisOfProcessing,
		PolicyURL:               pReq.PolicyURL,
		AttributeType:           pReq.AttributeType,
		Jurisdiction:            pReq.Jurisdiction,
		Disclosure:              pReq.Disclosure,
		IndustryScope:           pReq.IndustryScope,
		DataRetention:           pReq.DataRetention,
		Restriction:             pReq.Restriction,
		Shared3PP:               pReq.Shared3PP,
		SSIID:                   pReq.SSIID,
	}
	o.Purposes = append(o.Purposes, newPurpose)

	_, err = org.UpdatePurposes(o.ID.Hex(), o.Purposes)
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
	response, _ := json.Marshal(newPurpose)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
