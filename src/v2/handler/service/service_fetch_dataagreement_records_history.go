package service

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/common"
	daRecordHistory "github.com/bb-consent/api/src/v2/dataagreement_record_history"
	"github.com/bb-consent/api/src/v2/paginate"
	"go.mongodb.org/mongo-driver/bson"
)

type listDataAgreementRecordHistory struct {
	DataAgreementRecordHistory interface{}         `json:"dataAgreementRecordHistory"`
	Pagination                 paginate.Pagination `json:"pagination"`
}

func ServiceFetchRecordsHistory(w http.ResponseWriter, r *http.Request) {
	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

	// Return all data agreement record histories
	var darH []daRecordHistory.DataAgreementRecordsHistory
	query := paginate.PaginateDBObjectsQuery{
		Filter:     bson.M{},
		Collection: daRecordHistory.Collection(),
		Context:    context.Background(),
		Limit:      limit,
		Offset:     offset,
	}
	var resp listDataAgreementRecordHistory
	result, err := paginate.PaginateDBObjects(query, &darH)
	if err != nil {
		if errors.Is(err, paginate.EmptyDBError) {
			emptyDarH := make([]interface{}, 0)
			resp = listDataAgreementRecordHistory{
				DataAgreementRecordHistory: emptyDarH,
				Pagination:                 result.Pagination,
			}
			common.ReturnHTTPResponse(resp, w)
			return
		}
		m := "Failed to paginate data agreement record histories"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return

	}
	resp = listDataAgreementRecordHistory{
		DataAgreementRecordHistory: result.Items,
		Pagination:                 result.Pagination,
	}
	common.ReturnHTTPResponse(resp, w)
}
