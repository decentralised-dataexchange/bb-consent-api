package dataagreement

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	"github.com/bb-consent/api/internal/paginate"
	"github.com/bb-consent/api/internal/policy"
	"github.com/bb-consent/api/internal/revision"
	"go.mongodb.org/mongo-driver/bson"
)

// ListDataAgreementsError is an error enumeration for list data agreement API.
type ListDataAgreementsError int

const (
	// ErrRevisionIDIsMissing indicates that the revisionId query param is missing.
	RevisionIDIsMissingError ListDataAgreementsError = iota
	LifecycleIsMissingError
)

// Error returns the string representation of the error.
func (e ListDataAgreementsError) Error() string {
	switch e {
	case RevisionIDIsMissingError:
		return "Query param revisionId is missing!"
	case LifecycleIsMissingError:
		return "Query param lifecycle is missing!"
	default:
		return "Unknown error!"
	}
}

// ParseListDataAgreementsQueryParams parses query params for listing data agreements.
func ParseListDataAgreementsQueryParams(r *http.Request) (revisionId string, err error) {
	query := r.URL.Query()

	// Check if revisionId query param is provided.
	if r, ok := query["revisionId"]; ok && len(r) > 0 {
		return r[0], nil
	}

	return "", RevisionIDIsMissingError
}

// ParseListDataAgreementsLifecycleQueryParams parses query params for listing data agreements.
func ParseListDataAgreementsLifecycleQueryParams(r *http.Request) (lifecycle string, err error) {
	query := r.URL.Query()

	// Check if revisionId query param is provided.
	if r, ok := query["lifecycle"]; ok && len(r) > 0 {
		return r[0], nil
	}

	return "", LifecycleIsMissingError
}

func parseIncludeRevisionsQueryParams(r *http.Request) (includeRevisions bool, err error) {
	// Check if includeRevisions query param is provided.
	includeRevisionsString := r.URL.Query().Get("includeRevisions")
	if len(strings.TrimSpace(includeRevisionsString)) >= 1 {
		includeRevisions, err := strconv.ParseBool(includeRevisionsString)
		if err != nil {
			return false, err
		}
		return includeRevisions, nil
	}

	return false, nil
}

func activeDataAgreementsFromObjectData(organisationId string) ([]interface{}, error) {
	var activeDataAgreements []interface{}
	dataAgreements, err := dataagreement.GetAllDataAgreementsWithLatestRevisionsObjectData(organisationId)
	if err != nil {
		return activeDataAgreements, err
	}

	for _, dataAgreement := range dataAgreements {
		if len(dataAgreement.ObjectData) >= 1 {
			// Recreate data agreement from revision
			activeDataAgreement, err := revision.RecreateDataAgreementFromObjectData(dataAgreement.ObjectData)
			if err != nil {
				return activeDataAgreements, err
			}
			activeDataAgreements = append(activeDataAgreements, activeDataAgreement)
		}
	}

	return activeDataAgreements, nil
}

// list data agreements based on lifecycle
func listDataAgreementsBasedOnLifecycle(lifecycle string, organisationId string) ([]interface{}, error) {
	darepo := dataagreement.DataAgreementRepository{}
	darepo.Init(organisationId)

	var dataAgreements []interface{}
	var err error

	switch lifecycle {
	case config.Complete:
		dataAgreements, err = activeDataAgreementsFromObjectData(organisationId)
		if err != nil {
			return dataAgreements, err
		}
	case config.Draft:
		draftDataAgreements, err := darepo.GetDataAgreementsByLifecycle(lifecycle)
		if err != nil {
			return dataAgreements, err
		}
		dataAgreements = dataAgreementsToInterfaceSlice(draftDataAgreements)

	}
	return dataAgreements, nil
}

func dataAgreementsToInterfaceSlice(dataAgreements []dataagreement.DataAgreement) []interface{} {
	interfaceSlice := make([]interface{}, len(dataAgreements))
	for i, r := range dataAgreements {
		interfaceSlice[i] = r
	}
	return interfaceSlice
}

func setDataAgreementWithRevisions(dataAgreement dataagreement.DataAgreement, revisions []revision.Revision) dataAgreementWithRevisions {
	var dataAgreementWithRevision dataAgreementWithRevisions
	// Policy
	dataAgreementWithRevision.Policy = dataAgreement.Policy

	// Signature
	dataAgreementWithRevision.Signature = dataAgreement.Signature

	// Other details
	dataAgreementWithRevision.Purpose = dataAgreement.Purpose
	dataAgreementWithRevision.PurposeDescription = dataAgreement.PurposeDescription
	dataAgreementWithRevision.LawfulBasis = dataAgreement.LawfulBasis
	dataAgreementWithRevision.MethodOfUse = dataAgreement.MethodOfUse
	dataAgreementWithRevision.DpiaDate = dataAgreement.DpiaDate
	dataAgreementWithRevision.DpiaSummaryUrl = dataAgreement.DpiaSummaryUrl
	dataAgreementWithRevision.Active = dataAgreement.Active
	dataAgreementWithRevision.Forgettable = dataAgreement.Forgettable
	dataAgreementWithRevision.CompatibleWithVersionId = dataAgreement.CompatibleWithVersionId
	dataAgreementWithRevision.DataAttributes = dataAgreement.DataAttributes
	dataAgreementWithRevision.Lifecycle = dataAgreement.Lifecycle
	dataAgreementWithRevision.Id = dataAgreement.Id
	dataAgreementWithRevision.ControllerId = dataAgreement.ControllerId
	dataAgreementWithRevision.ControllerName = dataAgreement.ControllerName
	dataAgreementWithRevision.ControllerUrl = dataAgreement.ControllerUrl
	dataAgreementWithRevision.Version = dataAgreement.Version
	dataAgreementWithRevision.Revisions = revisions
	dataAgreementWithRevision.Controller = dataAgreement.Controller
	dataAgreementWithRevision.Dpia = dataAgreement.Dpia
	dataAgreementWithRevision.DataUse = dataAgreement.DataUse
	dataAgreementWithRevision.CompatibleWithVersion = dataAgreement.CompatibleWithVersion
	dataAgreementWithRevision.DataSources = dataAgreement.DataSources

	return dataAgreementWithRevision
}

func getDataAgreementsWithRevisions(organisationId string, lifecycle string) ([]dataAgreementWithRevisions, error) {
	darepo := dataagreement.DataAgreementRepository{}
	darepo.Init(organisationId)

	var tempDataAgreements []dataAgreementWithRevisions
	var dataAgreements []dataagreement.DataAgreement
	var err error

	dataAgreements, err = darepo.GetAll()
	if err != nil {
		return tempDataAgreements, err
	}

	for _, dataAgreement := range dataAgreements {
		// list all revisions for data agreement
		revisions, err := revision.ListAllByObjectIdAndSchemaName(dataAgreement.Id, config.DataAgreement)
		if err != nil {
			return tempDataAgreements, err
		}
		if len(strings.TrimSpace(lifecycle)) > 1 {
			// if lifecycle query param is present, return data agreements filtered by lifecycle
			switch lifecycle {
			case config.Complete:
				// remove the draft data agreements by checking if revision is present
				if len(revisions) >= 1 {
					dAFromRevision, err := revision.RecreateDataAgreementFromRevision(revisions[0])
					if err != nil {
						return tempDataAgreements, err
					}
					tempDataAgreement := setDataAgreementWithRevisions(dAFromRevision, revisions)
					tempDataAgreements = append(tempDataAgreements, tempDataAgreement)
				}
			case config.Draft:
				// remove the data agreements which are active
				if !dataAgreement.Active {
					tempDataAgreement := setDataAgreementWithRevisions(dataAgreement, revisions)
					tempDataAgreements = append(tempDataAgreements, tempDataAgreement)
				}

			}

		} else {
			// return data agreements with out lifecycle filtered
			tempDataAgreement := setDataAgreementWithRevisions(dataAgreement, revisions)
			tempDataAgreements = append(tempDataAgreements, tempDataAgreement)
		}

	}
	return tempDataAgreements, nil
}

type dataAgreementWithRevisions struct {
	Id                      string                        `json:"id" bson:"_id,omitempty"`
	Version                 string                        `json:"version"`
	ControllerId            string                        `json:"controllerId"`
	ControllerUrl           string                        `json:"controllerUrl" valid:"required"`
	ControllerName          string                        `json:"controllerName" valid:"required"`
	Policy                  policy.Policy                 `json:"policy" valid:"required"`
	Purpose                 string                        `json:"purpose" valid:"required"`
	PurposeDescription      string                        `json:"purposeDescription" valid:"required"`
	LawfulBasis             string                        `json:"lawfulBasis" valid:"required"`
	MethodOfUse             string                        `json:"methodOfUse" valid:"required"`
	DpiaDate                string                        `json:"dpiaDate"`
	DpiaSummaryUrl          string                        `json:"dpiaSummaryUrl"`
	Signature               dataagreement.Signature       `json:"signature"`
	Active                  bool                          `json:"active"`
	Forgettable             bool                          `json:"forgettable"`
	CompatibleWithVersionId string                        `json:"compatibleWithVersionId"`
	Lifecycle               string                        `json:"lifecycle" valid:"required"`
	DataAttributes          []dataagreement.DataAttribute `json:"dataAttributes" valid:"required"`
	OrganisationId          string                        `json:"-"`
	IsDeleted               bool                          `json:"-"`
	Timestamp               string                        `json:"-"`
	Revisions               []revision.Revision           `json:"revisions"`
	DataUse                 string                        `json:"dataUse"`
	Dpia                    string                        `json:"dpia"`
	CompatibleWithVersion   string                        `json:"compatibleWithVersion"`
	Controller              dataagreement.Controller      `json:"controller"`
	DataSources             []dataagreement.DataSource    `json:"dataSources"`
}

func dataAgreementsWithRevisionsToInterfaceSlice(dataAgreements []dataAgreementWithRevisions) []interface{} {
	interfaceSlice := make([]interface{}, len(dataAgreements))
	for i, r := range dataAgreements {
		interfaceSlice[i] = r
	}
	return interfaceSlice
}

type listDataAgreementsResp struct {
	DataAgreements interface{}         `json:"dataAgreements"`
	Pagination     paginate.Pagination `json:"pagination"`
}

func returnHTTPResponse(resp interface{}, w http.ResponseWriter) {
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}

// ConfigListDataAgreements
func ConfigListDataAgreements(w http.ResponseWriter, r *http.Request) {
	// Headers
	organisationId := r.Header.Get(config.OrganizationId)
	organisationId = common.Sanitize(organisationId)

	var resp listDataAgreementsResp

	// Query params
	offset, limit := paginate.ParsePaginationQueryParams(r)
	log.Printf("Offset: %v and limit: %v\n", offset, limit)

	includeRevisions, err := parseIncludeRevisionsQueryParams(r)
	if err != nil {
		m := "Failed to parse query param includeRevisions"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	revisionId, err := ParseListDataAgreementsQueryParams(r)
	revisionId = common.Sanitize(revisionId)
	if err != nil && errors.Is(err, RevisionIDIsMissingError) {
		lifecycle, err := ParseListDataAgreementsLifecycleQueryParams(r)
		lifecycle = common.Sanitize(lifecycle)

		darepo := dataagreement.DataAgreementRepository{}
		darepo.Init(organisationId)

		if err != nil && errors.Is(err, LifecycleIsMissingError) {

			if includeRevisions {
				dataAgreements, err := getDataAgreementsWithRevisions(organisationId, lifecycle)
				if err != nil {
					m := "Failed to fetch data agreements"
					common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
					return
				}

				interfaceSlice := dataAgreementsWithRevisionsToInterfaceSlice(dataAgreements)

				// Return liecycle filtered data agreements
				query := paginate.PaginateObjectsQuery{
					Limit:  limit,
					Offset: offset,
				}
				result := paginate.PaginateObjects(query, interfaceSlice)

				resp = listDataAgreementsResp{
					DataAgreements: result.Items,
					Pagination:     result.Pagination,
				}
				common.ReturnHTTPResponse(resp, w)
				return
			}

			var pipeline []bson.M
			pipeline = append(pipeline, bson.M{"$match": bson.M{"organisationid": organisationId, "isdeleted": false}})
			pipeline = append(pipeline, bson.M{"$sort": bson.M{"timestamp": -1}})
			// Return all data agreements
			var dataAgreements []dataagreement.DataAgreement

			query := paginate.PaginateDBObjectsQueryUsingPipeline{
				Pipeline:   pipeline,
				Collection: dataagreement.Collection(),
				Context:    context.Background(),
				Limit:      limit,
				Offset:     offset,
			}
			result, err := paginate.PaginateDBObjectsUsingPipeline(query, &dataAgreements)
			if err != nil {
				if errors.Is(err, paginate.EmptyDBError) {
					emptyDataAgreements := make([]interface{}, 0)
					resp = listDataAgreementsResp{
						DataAgreements: emptyDataAgreements,
						Pagination:     result.Pagination,
					}
					returnHTTPResponse(resp, w)
					return
				}
				m := "Failed to paginate data agreement"
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return

			}
			resp = listDataAgreementsResp{
				DataAgreements: result.Items,
				Pagination:     result.Pagination,
			}
			returnHTTPResponse(resp, w)
			return
		} else {
			// Both lifecycle and includeRevisions are present
			if includeRevisions {
				dataAgreements, err := getDataAgreementsWithRevisions(organisationId, lifecycle)
				if err != nil {
					m := "Failed to fetch data agreements"
					common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
					return
				}

				interfaceSlice := dataAgreementsWithRevisionsToInterfaceSlice(dataAgreements)

				// Return liecycle filtered data agreements
				query := paginate.PaginateObjectsQuery{
					Limit:  limit,
					Offset: offset,
				}
				result := paginate.PaginateObjects(query, interfaceSlice)

				resp = listDataAgreementsResp{
					DataAgreements: result.Items,
					Pagination:     result.Pagination,
				}
				common.ReturnHTTPResponse(resp, w)
				return
			}

			// if lifecycle is present
			dataAgreements, err := listDataAgreementsBasedOnLifecycle(lifecycle, organisationId)
			if err != nil {
				m := "Failed to fetch data agreements"
				common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
				return
			}

			// Return liecycle filtered data agreements
			query := paginate.PaginateObjectsQuery{
				Limit:  limit,
				Offset: offset,
			}
			result := paginate.PaginateObjects(query, dataAgreements)

			resp = listDataAgreementsResp{
				DataAgreements: result.Items,
				Pagination:     result.Pagination,
			}
			common.ReturnHTTPResponse(resp, w)
			return

		}

	} else {
		// Fetch revision by id
		revisionResp, err := revision.GetByRevisionId(revisionId)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch revision by id: %v", revisionId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

		// Recreate data agreement from revision
		da, err := revision.RecreateDataAgreementFromRevision(revisionResp)
		if err != nil {
			m := fmt.Sprintf("Failed to fetch data agreement by revision: %v", revisionId)
			common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
			return
		}

		interfaceSlice := make([]interface{}, 0)
		interfaceSlice = append(interfaceSlice, da)

		// Constructing the response
		resp = listDataAgreementsResp{
			DataAgreements: interfaceSlice,
			Pagination: paginate.Pagination{
				CurrentPage: 1,
				TotalItems:  1,
				TotalPages:  1,
				Limit:       1,
				HasPrevious: false,
				HasNext:     false,
			},
		}

	}

	returnHTTPResponse(resp, w)
}
