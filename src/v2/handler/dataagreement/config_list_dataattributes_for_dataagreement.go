package dataagreement

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/dataagreement"
	"github.com/bb-consent/api/src/dataattribute"
	"github.com/bb-consent/api/src/paginate"
	"github.com/gorilla/mux"
)

type listDataAttributesResp struct {
	DataAgreement  dataagreement.DataAgreement `json:"dataAgreement"`
	DataAttributes interface{}                 `json:"dataAttributes"`
	Pagination     paginate.Pagination         `json:"pagination"`
}

func dataAttributesToInterfaceSlice(dataAttributes []dataattribute.DataAttribute) []interface{} {
	interfaceSlice := make([]interface{}, len(dataAttributes))
	for i, r := range dataAttributes {
		interfaceSlice[i] = r
	}
	return interfaceSlice
}

func ConfigListDataAttributesForDataAgreement(w http.ResponseWriter, r *http.Request) {
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	dataAgreementId := mux.Vars(r)[config.DataAgreementId]
	dataAgreementId = common.Sanitize(dataAgreementId)

	// Repository
	daRepo := dataagreement.DataAgreementRepository{}
	daRepo.Init(organisationId)

	dataAttributeRepo := dataattribute.DataAttributeRepository{}
	dataAttributeRepo.Init(organisationId)

	da, err := daRepo.Get(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	dataAttributes, err := dataAttributeRepo.GetDataAttributesByDataAgreementId(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data attributes for data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

	query := paginate.PaginateObjectsQuery{
		Limit:  limit,
		Offset: offset,
	}
	interfaceSlice := dataAttributesToInterfaceSlice(dataAttributes)
	result := paginate.PaginateObjects(query, interfaceSlice)

	var resp = listDataAttributesResp{
		DataAgreement:  da,
		DataAttributes: result.Items,
		Pagination:     result.Pagination,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
