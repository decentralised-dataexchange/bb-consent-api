package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/paginate"
	"github.com/bb-consent/api/internal/revision"
	"github.com/gorilla/mux"
)

type listDataAttributesResp struct {
	DataAgreement  dataagreement.DataAgreement `json:"dataAgreement"`
	DataAttributes interface{}                 `json:"dataAttributes"`
	Pagination     paginate.Pagination         `json:"pagination"`
}

func dataAttributesToInterfaceSlice(dataAttributes []dataagreement.DataAttribute) []interface{} {
	interfaceSlice := make([]interface{}, len(dataAttributes))
	for i, r := range dataAttributes {
		interfaceSlice[i] = r
	}
	return interfaceSlice
}

func ServiceListDataAttributesForDataAgreement(w http.ResponseWriter, r *http.Request) {
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

	revisionId, err := ParseListDataAgreementsQueryParams(r)
	revisionId = common.Sanitize(revisionId)
	var daRevision revision.Revision
	if err != nil && errors.Is(err, RevisionIDIsMissingError) {

		daRevision, err = revision.GetLatestByDataAgreementId(da.Id.Hex())
		if err != nil {
			m := fmt.Sprintf("Failed to fetch data agreement revision: %v", dataAgreementId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
	} else {
		daRevision, err = revision.GetByRevisionIdAndSchema(revisionId, config.DataAgreement)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch data agreement revision: %v", dataAgreementId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}
	}

	dataAgreement, err := revision.RecreateDataAgreementFromRevision(daRevision)
	if err != nil {
		m := fmt.Sprintf("Failed to recreate data agreement from revision: %v", dataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	dataAttributes := dataAgreement.DataAttributes

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
		DataAgreement:  dataAgreement,
		DataAttributes: result.Items,
		Pagination:     result.Pagination,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
