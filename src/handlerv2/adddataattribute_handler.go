package handlerv2

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

type template struct {
	Consent    string   `valid:"required"`
	PurposeIDs []string `valid:"required"`
}
type templateReq struct {
	Templates []template
}

// AddDataAttribute Adds an organization data attribute
func AddDataAttribute(w http.ResponseWriter, r *http.Request) {
	organizationID := r.Header.Get(config.OrganizationId)

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
			ID:         primitive.NewObjectID().Hex(),
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
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
