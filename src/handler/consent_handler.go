package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/consent"
	"github.com/bb-consent/api/src/consenthistory"
	"github.com/bb-consent/api/src/org"
	"github.com/gorilla/mux"
)

type consentStatus struct {
	Consented string
	TimeStamp time.Time
	Days      int
	Remaining int
}

// ConsentCount Counts the total consent attributes and consented ones
type ConsentCount struct {
	Total     int
	Consented int
}
type consentResp struct {
	ID          string
	Description string
	Value       string
	Status      consentStatus
}
type consentsAndPurpose struct {
	Purpose  org.Purpose
	Count    ConsentCount
	Consents []consentResp
}

// ConsentsResp Consent response struct definition
type ConsentsResp struct {
	ID                  string
	OrgID               string
	UserID              string
	ConsentsAndPurposes []consentsAndPurpose
}

// DataRetentionPolicyResp Data retention policy response struct definition
type DataRetentionPolicyResp struct {
	Expiry string
}

type consentsAndPurposeWithDataRetention struct {
	Purpose       org.Purpose
	Count         ConsentCount
	Consents      []consentResp
	DataRetention DataRetentionPolicyResp
}

// ConsentsRespWithDataRetention Consent response struct definition with data retention for each purpose
type ConsentsRespWithDataRetention struct {
	ID                  string
	OrgID               string
	UserID              string
	ConsentsAndPurposes []consentsAndPurposeWithDataRetention
}

// GetConsentResponse Gets all consents and formulates the response
func GetConsentResponse(w http.ResponseWriter, userID string, orgID string) (ConsentsResp, error) {
	var c ConsentsResp
	o, err := org.Get(orgID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization user: %v org: %v", userID, orgID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return c, err
	}

	consents, err := consent.GetByUserOrg(userID, orgID)
	if err != nil {
		if err.Error() == "not found" {
			var con consent.Consents
			con.OrgID = orgID
			con.UserID = userID
			consents, err = consent.Add(con)
			if err != nil {
				m := fmt.Sprintf("Failed to fetch consents user: %v org: %v", userID, orgID)
				common.HandleError(w, http.StatusInternalServerError, m, err)
				return c, err
			}
		} else {
			m := fmt.Sprintf("Failed to fetch consents user: %v org: %v", userID, orgID)
			common.HandleError(w, http.StatusInternalServerError, m, err)
			return c, err
		}
	}
	c = createConsentGetResponse(consents, o)
	c.ID = consents.ID.Hex()
	c.OrgID = orgID
	c.UserID = userID

	return c, nil
}

// GetConsents Gets all consent entries in the collection
func GetConsents(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgID"]
	userID := mux.Vars(r)["userID"]

	// Fetching the organisation by ID
	o, err := org.Get(orgID)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch organization for user: %v org: %v", userID, orgID)
		common.HandleError(w, http.StatusInternalServerError, m, err)
		return
	}

	c, err := GetConsentResponse(w, userID, orgID)
	if err != nil {
		log.Printf("Failed to get consents for user: %v org: %v err: %v", userID, orgID, err)
		return
	}

	// For holding the API response
	var RespData ConsentsRespWithDataRetention

	// Constructing the response data
	RespData.ID = c.ID
	RespData.UserID = c.UserID
	RespData.OrgID = c.OrgID

	// Constructing the response data
	if len(c.ConsentsAndPurposes) > 0 {
		for _, tempConsentsAndPurpose := range c.ConsentsAndPurposes {

			var tempConsentsAndPurposes consentsAndPurposeWithDataRetention
			tempConsentsAndPurposes.Purpose = tempConsentsAndPurpose.Purpose
			tempConsentsAndPurposes.Count = tempConsentsAndPurpose.Count

			if tempConsentsAndPurpose.Count.Total > 0 {
				if tempConsentsAndPurpose.Consents == nil {
					tempConsentsAndPurposes.Consents = make([]consentResp, 0)
				} else {
					tempConsentsAndPurposes.Consents = tempConsentsAndPurpose.Consents
				}

				RespData.ConsentsAndPurposes = append(RespData.ConsentsAndPurposes, tempConsentsAndPurposes)
			}

		}

	}

	// Check if data retention policy enabled for the organisation
	if o.DataRetention.Enabled {

		if len(RespData.ConsentsAndPurposes) > 0 {
			// Add data retention expiry for each purpose if available
			for i, _ := range RespData.ConsentsAndPurposes {

				// Check if purpose has consent as lawful basis
				if !RespData.ConsentsAndPurposes[i].Purpose.LawfulUsage {

					// Check if the purpose is allowed
					if RespData.ConsentsAndPurposes[i].Count.Consented > 0 {

						latestConsentHistory, err := consenthistory.GetLatestByUserOrgPurposeID(RespData.UserID, RespData.OrgID, RespData.ConsentsAndPurposes[i].Purpose.ID)
						if err != nil {
							continue
						}

						RespData.ConsentsAndPurposes[i].DataRetention.Expiry = latestConsentHistory.ID.Time().Add(time.Second * time.Duration(o.DataRetention.RetentionPeriod)).UTC().String()
						log.Printf("Expiry for purpose:%v is %v", RespData.ConsentsAndPurposes[i].Purpose.ID, RespData.ConsentsAndPurposes[i].DataRetention.Expiry)
					}

				}

			}
		}

	}

	//fmt.Printf("c:%v", c)
	response, _ := json.Marshal(RespData)

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
