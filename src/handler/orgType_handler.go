package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/src/common"
	ot "github.com/bb-consent/api/src/orgtype"
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
