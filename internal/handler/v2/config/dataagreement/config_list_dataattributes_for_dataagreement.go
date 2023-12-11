package dataagreement

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/paginate"
	"github.com/gorilla/mux"
)

type dataAgreementForDataAttribute struct {
	Id      string `json:"id" bson:"_id,omitempty"`
	Purpose string `json:"purpose"`
}

type dataAttributeForLists struct {
	Id            string                        `json:"id" bson:"_id,omitempty"`
	Name          string                        `json:"name" valid:"required"`
	Description   string                        `json:"description" valid:"required"`
	Sensitivity   bool                          `json:"sensitivity"`
	Category      string                        `json:"category"`
	DataAgreement dataAgreementForDataAttribute `json:"dataAgreement"`
}

type listDataAttributesResp struct {
	DataAttributes interface{}         `json:"dataAttributes"`
	Pagination     paginate.Pagination `json:"pagination"`
}

func dataAttributesForList(dA dataagreement.DataAgreement) []dataAttributeForLists {

	var dataAttributes []dataAttributeForLists

	for _, dataAttribute := range dA.DataAttributes {
		var tempDataAttribute dataAttributeForLists
		tempDataAttribute.Id = dataAttribute.Id
		tempDataAttribute.Name = dataAttribute.Name
		tempDataAttribute.Description = dataAttribute.Description
		tempDataAttribute.Sensitivity = dataAttribute.Sensitivity
		tempDataAttribute.Category = dataAttribute.Category
		tempDataAttribute.DataAgreement.Id = dA.Id
		tempDataAttribute.DataAgreement.Purpose = dA.Purpose
		dataAttributes = append(dataAttributes, tempDataAttribute)

	}

	return dataAttributes
}

func dataAttributesToInterfaceSlice(dataAttributes []dataAttributeForLists) []interface{} {
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

	da, err := daRepo.Get(dataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch data agreement: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	dataAttributes := dataAttributesForList(da)

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
		DataAttributes: result.Items,
		Pagination:     result.Pagination,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
