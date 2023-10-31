package common

import (
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/exp/maps"
)

func CombineFilters(filter1 bson.M, filter2 bson.M) bson.M {
	maps.Copy(filter1, filter2)
	return filter1
}
