package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/org"
	"github.com/bb-consent/api/src/token"
)

type userOrg struct {
	ID            string
	Name          string
	CoverImageURL string
	LogoImageURL  string
	Location      string
	TypeID        string
	Type          string
	PolicyURL     string
	Description   string
}

func convertToUserOrg(o org.Organization) (u userOrg) {
	var uo userOrg
	uo.ID = o.ID.Hex()
	uo.Name = o.Name
	uo.CoverImageURL = o.CoverImageURL
	uo.LogoImageURL = o.LogoImageURL
	uo.Location = o.Location
	uo.TypeID = o.Type.ID.Hex()
	uo.Type = o.Type.Type
	uo.PolicyURL = o.PolicyURL
	uo.Description = o.Description

	return uo
}

type orgConsents struct {
	Organization    userOrg
	ConsentID       string
	PurposeConsents []purposeConsentsBrief
}

type purposeConsentsBrief struct {
	Purpose org.Purpose
	Count   ConsentCount
}

func transformConsentResponse(c ConsentsResp) []purposeConsentsBrief {
	var pc []purposeConsentsBrief

	for _, item := range c.ConsentsAndPurposes {
		var p purposeConsentsBrief
		p.Count = item.Count
		p.Purpose = item.Purpose

		pc = append(pc, p)
	}

	return pc
}

// GetUserOrgsAndConsents Get org details and all consents
func GetUserOrgsAndConsents(w http.ResponseWriter, r *http.Request) {
	userID := token.GetUserID(r)
	orgIDs, ok := r.URL.Query()["orgID"]

	if !ok || len(orgIDs) < 1 {
		m := fmt.Sprintf("Missing type query parameter orgID for userID: %v", userID)
		common.HandleError(w, http.StatusBadRequest, m, nil)
		return
	}

	orgID := orgIDs[0]

	org, err := org.Get(orgID)
	if err != nil {
		m := fmt.Sprintf("Failed to get organization: %v", orgID)
		common.HandleError(w, http.StatusNotFound, m, err)
		return
	}

	consents, err := GetConsentResponse(w, userID, orgID)
	if err != nil {
		log.Printf("Failed to get consents for user: %v org: %v err: %v", userID, orgID, err)
		return
	}

	c := orgConsents{convertToUserOrg(org), consents.ID, transformConsentResponse(consents)}

	response, _ := json.Marshal(c)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
