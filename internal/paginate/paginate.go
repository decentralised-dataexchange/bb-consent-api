package paginate

import (
	"context"
	"errors"
	"log"
	"math"
	"net/http"
	"reflect"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Pagination
type Pagination struct {
	CurrentPage int  `json:"currentPage"`
	TotalItems  int  `json:"totalItems"`
	TotalPages  int  `json:"totalPages"`
	Limit       int  `json:"limit"`
	HasPrevious bool `json:"hasPrevious"`
	HasNext     bool `json:"hasNext"`
}

// PaginateDBObjectsQuery
type PaginateDBObjectsQuery struct {
	Filter     bson.M
	Collection *mongo.Collection
	Context    context.Context
	Limit      int
	Offset     int
}

// PaginateDBObjectsQuery
type PaginateDBObjectsQueryUsingPipeline struct {
	Pipeline   []bson.M
	Collection *mongo.Collection
	Context    context.Context
	Limit      int
	Offset     int
}

// PaginatedDBResult
type PaginatedDBResult struct {
	Items      interface{} `json:"items"`
	Pagination Pagination  `json:"pagination"`
}

// PaginationError is an error enumeration for pagination
type PaginationError int

const (
	// EmptyDBError indicates that database is empty.
	EmptyDBError PaginationError = iota
)

// Error returns the string representation of the error.
func (e PaginationError) Error() string {
	switch e {
	case EmptyDBError:
		return "Database is empty!"
	default:
		return "Unknown error!"
	}
}

// PaginateDBObjects
func PaginateDBObjects(query PaginateDBObjectsQuery, resultSlice interface{}) (*PaginatedDBResult, error) {

	// Calculate total items
	totalItems, err := query.Collection.CountDocuments(query.Context, query.Filter)
	if err != nil {
		return nil, err
	}

	if totalItems == 0 {
		return &PaginatedDBResult{}, EmptyDBError
	}

	// Ensure offset is not negative and limit is positive
	if query.Offset < 0 {
		query.Offset = 0
	}
	if query.Limit <= 0 {
		query.Limit = 1
	}

	// Ensure offset is within bounds
	if query.Offset >= int(totalItems) {
		query.Offset = int(totalItems) - query.Limit
	}

	// Calculate pages and selected page based on offset and limit
	totalPages := int(math.Ceil(float64(totalItems) / float64(query.Limit)))
	currentPage := (query.Offset / query.Limit) + 1

	// Ensure currentPage is within bounds
	if currentPage < 1 {
		currentPage = 1
	} else if currentPage > totalPages {
		currentPage = totalPages
	}

	// Initialize pagination structure
	pagination := Pagination{
		CurrentPage: currentPage,
		TotalItems:  int(totalItems),
		Limit:       query.Limit,
		TotalPages:  totalPages,
		HasPrevious: currentPage > 1,
		HasNext:     currentPage < totalPages,
	}

	// Query the database
	opts := options.Find().SetSkip(int64(query.Offset)).SetLimit(int64(query.Limit))
	cursor, err := query.Collection.Find(query.Context, query.Filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(query.Context)

	// Decode items
	sliceValue := reflect.ValueOf(resultSlice)
	if sliceValue.Kind() != reflect.Ptr || sliceValue.Elem().Kind() != reflect.Slice {
		return nil, errors.New("resultSlice must be a slice pointer")
	}
	sliceElem := sliceValue.Elem()
	itemTyp := sliceElem.Type().Elem()

	for cursor.Next(query.Context) {
		itemPtr := reflect.New(itemTyp).Interface()
		if err := cursor.Decode(itemPtr); err != nil {
			return nil, err
		}
		sliceElem = reflect.Append(sliceElem, reflect.ValueOf(itemPtr).Elem())
	}
	if sliceElem.Len() == 0 {
		sliceElem = reflect.MakeSlice(sliceElem.Type(), 0, 0)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return &PaginatedDBResult{
		Items:      sliceElem.Interface(),
		Pagination: pagination,
	}, nil
}

// PaginateObjectsQuery
type PaginateObjectsQuery struct {
	Limit  int
	Offset int
}

// PaginatedResult
type PaginatedResult struct {
	Items      interface{} `json:"items"`
	Pagination Pagination  `json:"pagination"`
}

func PaginateObjects(query PaginateObjectsQuery, toBeSortedItems []interface{}) *PaginatedResult {
	totalItems := len(toBeSortedItems)

	// Ensure offset is not negative and limit is positive
	if query.Offset < 0 {
		query.Offset = 0
	}
	if query.Limit <= 0 {
		query.Limit = 1
	}

	// Ensure offset is within bounds
	if query.Offset >= totalItems {
		query.Offset = totalItems - query.Limit
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	// Calculate pages and selected page based on offset and limit
	totalPages := int(math.Ceil(float64(totalItems) / float64(query.Limit)))
	currentPage := (query.Offset / query.Limit) + 1

	// Ensure currentPage is within bounds
	if currentPage < 1 {
		currentPage = 1
	} else if currentPage > totalPages {
		currentPage = totalPages
	}

	endIdx := query.Offset + query.Limit
	if endIdx > totalItems {
		endIdx = totalItems
	}

	paginatedItems := toBeSortedItems[query.Offset:endIdx]

	return &PaginatedResult{
		Items: paginatedItems,
		Pagination: Pagination{
			CurrentPage: currentPage,
			TotalItems:  totalItems,
			TotalPages:  totalPages,
			Limit:       query.Limit,
			HasPrevious: query.Offset > 0,
			HasNext:     endIdx < totalItems,
		},
	}

}

// PaginateDBObjects
func PaginateDBObjectsUsingPipeline(query PaginateDBObjectsQueryUsingPipeline, resultSlice interface{}) (*PaginatedDBResult, error) {

	// Calculate total items
	cursor, err := query.Collection.Aggregate(context.TODO(), query.Pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var dbObjectsForCount []interface{}
	err = cursor.All(context.TODO(), &dbObjectsForCount)
	if err != nil {
		return nil, err
	}

	totalItems := len(dbObjectsForCount)

	if totalItems == 0 {
		return &PaginatedDBResult{
			Items: []string{},
			Pagination: Pagination{
				CurrentPage: 1,
				TotalItems:  0,
				Limit:       query.Limit,
				TotalPages:  1,
				HasPrevious: false,
				HasNext:     false,
			},
		}, nil
	}

	// Ensure offset is not negative and limit is positive
	if query.Offset < 0 {
		query.Offset = 0
	}
	if query.Limit <= 0 {
		query.Limit = 1
	}

	log.Printf("Current totalItems: %d", totalItems)

	// Ensure offset is within bounds
	if query.Offset >= int(totalItems) {
		query.Offset = int(totalItems) - query.Limit
	}

	// Calculate pages and selected page based on offset and limit
	totalPages := int(math.Ceil(float64(totalItems) / float64(query.Limit)))
	currentPage := (query.Offset / query.Limit) + 1

	// Ensure currentPage is within bounds
	if currentPage < 1 {
		currentPage = 1
	} else if currentPage > totalPages {
		currentPage = totalPages
	}

	// Initialize pagination structure
	pagination := Pagination{
		CurrentPage: currentPage,
		TotalItems:  int(totalItems),
		Limit:       query.Limit,
		TotalPages:  totalPages,
		HasPrevious: currentPage > 1,
		HasNext:     currentPage < totalPages,
	}

	pipelineWithSkipAndLimitStages := append(query.Pipeline, bson.M{"$skip": query.Offset}, bson.M{"$limit": query.Limit})

	// Query the database
	cursor2, err := query.Collection.Aggregate(context.TODO(), pipelineWithSkipAndLimitStages)
	if err != nil {
		return nil, err
	}
	defer cursor2.Close(context.TODO())

	// Decode items
	sliceValue := reflect.ValueOf(resultSlice)
	if sliceValue.Kind() != reflect.Ptr || sliceValue.Elem().Kind() != reflect.Slice {
		return nil, errors.New("resultSlice must be a slice pointer")
	}
	sliceElem := sliceValue.Elem()
	itemTyp := sliceElem.Type().Elem()

	for cursor2.Next(query.Context) {
		itemPtr := reflect.New(itemTyp).Interface()
		if err := cursor2.Decode(itemPtr); err != nil {
			return nil, err
		}
		sliceElem = reflect.Append(sliceElem, reflect.ValueOf(itemPtr).Elem())
	}
	if sliceElem.Len() == 0 {
		sliceElem = reflect.MakeSlice(sliceElem.Type(), 0, 0)
	}

	if err := cursor2.Err(); err != nil {
		return nil, err
	}

	return &PaginatedDBResult{
		Items:      sliceElem.Interface(),
		Pagination: pagination,
	}, nil
}

const DEFAULT_LIMIT int = 10
const DEFAULT_OFFSET int = 0

// ParsePaginationQueryParams parses offset and limit from query parameters.
// If they are not available or invalid it returns default values.
func ParsePaginationQueryParams(r *http.Request) (offset int, limit int) {
	query := r.URL.Query()
	offset, limit = DEFAULT_OFFSET, DEFAULT_LIMIT

	// Check if offset query param is provided and if it is a valid integer.
	if o, ok := query["offset"]; ok && len(o) > 0 {
		if oInt, err := strconv.Atoi(o[0]); err == nil && oInt >= 0 {
			offset = oInt
		}
	}

	// Check if limit query param is provided and if it is a valid integer.
	if l, ok := query["limit"]; ok && len(l) > 0 {
		if lInt, err := strconv.Atoi(l[0]); err == nil && lInt > 0 {
			limit = lInt
		}
	}

	return offset, limit
}
