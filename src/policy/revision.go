package policy

import (
	"encoding/json"
	"time"

	"github.com/bb-consent/api/src/common"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Revision
type Revision struct {
	Id                       string `json:"id"`
	SchemaName               string `json:"schemaName"`
	ObjectId                 string `json:"objectId"`
	SignedWithoutObjectId    bool   `json:"signedWithoutObjectId"`
	Timestamp                string `json:"timestamp"`
	AuthorizedByIndividualId string `json:"authorizedByIndividualId"`
	AuthorizedByOtherId      string `json:"authorizedByOtherId"`
	PredecessorHash          string `json:"predecessorHash"`
	PredecessorSignature     string `json:"predecessorSignature"`
	ObjectData               string `json:"objectData"`
	SuccessorId              string `json:"-"`
	SerializedHash           string `json:"-"`
	SerializedSnapshot       string `json:"-"`
}

// Init
func (r *Revision) Init(objectId string, authorisedByOtherId string) {
	r.Id = primitive.NewObjectID().Hex()
	r.SchemaName = "policy"
	r.ObjectId = objectId
	r.SignedWithoutObjectId = false
	r.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	r.AuthorizedByIndividualId = ""
	r.AuthorizedByOtherId = authorisedByOtherId
}

func (r *Revision) updateSuccessorId(successorId string) {
	r.SuccessorId = successorId
}

func (r *Revision) updatePredecessorSignature(signature string) {
	r.PredecessorSignature = signature
}

func (r *Revision) updatePredecessorHash(hash string) {
	r.PredecessorHash = hash
}

// CreateRevision
func (r *Revision) CreateRevision(objectData interface{}) error {

	// Object data
	objectDataSerialised, err := json.Marshal(objectData)
	if err != nil {
		return err
	}
	r.ObjectData = string(objectDataSerialised)

	// Serialised snapshot
	// TODO: Use a standard json normalisation algorithm for .e.g JCS
	serialisedSnapshot, err := json.Marshal(r)
	if err != nil {
		return err
	}
	r.SerializedSnapshot = string(serialisedSnapshot)

	// Serialised hash using SHA-1
	r.SerializedHash, err = common.CalculateSHA1(string(serialisedSnapshot))
	if err != nil {
		return err
	}

	return nil

}

// UpdateRevision
func (r *Revision) UpdateRevision(previousRevision Revision, objectData interface{}) error {
	// Update successor for previous revision
	previousRevision.updateSuccessorId(r.Id)

	// Predecessor hash
	r.updatePredecessorHash(previousRevision.PredecessorHash)

	// Predecessor signature
	// TODO: Add signature for predecessor hash
	signature := ""
	r.updatePredecessorSignature(signature)

	// Create revision
	err := r.CreateRevision(objectData)
	if err != nil {
		return err
	}

	return nil
}

type policyForObjectData struct {
	Id                      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name                    string             `json:"name" valid:"required"`
	Version                 string             `json:"version"`
	Url                     string             `json:"url" valid:"required"`
	Jurisdiction            string             `json:"jurisdiction"`
	IndustrySector          string             `json:"industrySector"`
	DataRetentionPeriodDays int                `json:"dataRetentionPeriod"`
	GeographicRestriction   string             `json:"geographicRestriction"`
	StorageLocation         string             `json:"storageLocation"`
}

// CreateRevisionForPolicy
func CreateRevisionForPolicy(newPolicy Policy, orgAdminId string) (Revision, error) {
	// Object data
	objectData := policyForObjectData{
		Id:                      newPolicy.Id,
		Name:                    newPolicy.Name,
		Version:                 newPolicy.Version,
		Url:                     newPolicy.Url,
		Jurisdiction:            newPolicy.Jurisdiction,
		IndustrySector:          newPolicy.IndustrySector,
		DataRetentionPeriodDays: newPolicy.DataRetentionPeriodDays,
		GeographicRestriction:   newPolicy.GeographicRestriction,
		StorageLocation:         newPolicy.StorageLocation,
	}

	// Create revision
	revision := Revision{}
	revision.Init(objectData.Id.Hex(), orgAdminId)
	err := revision.CreateRevision(objectData)

	return revision, err
}

// UpdateRevisionForPolicy
func UpdateRevisionForPolicy(updatedPolicy Policy, previousRevision Revision, orgAdminId string) (Revision, error) {
	// Object data
	objectData := policyForObjectData{
		Id:                      updatedPolicy.Id,
		Name:                    updatedPolicy.Name,
		Version:                 updatedPolicy.Version,
		Url:                     updatedPolicy.Url,
		Jurisdiction:            updatedPolicy.Jurisdiction,
		IndustrySector:          updatedPolicy.IndustrySector,
		DataRetentionPeriodDays: updatedPolicy.DataRetentionPeriodDays,
		GeographicRestriction:   updatedPolicy.GeographicRestriction,
		StorageLocation:         updatedPolicy.StorageLocation,
	}

	// Update revision
	revision := Revision{}
	revision.Init(objectData.Id.Hex(), orgAdminId)
	err := revision.UpdateRevision(previousRevision, objectData)

	return revision, err
}

func RecreatePolicyFromRevision(revision Revision) (Policy, error) {

	// Deserialise revision snapshot
	var r Revision
	err := json.Unmarshal([]byte(revision.SerializedSnapshot), &r)
	if err != nil {
		return Policy{}, err
	}

	// Deserialise policy
	var p Policy
	err = json.Unmarshal([]byte(r.ObjectData), &p)
	if err != nil {
		return Policy{}, err
	}

	return p, nil
}

// RevisionForHTTPResponse
type RevisionForHTTPResponse struct {
	Revision
	SuccessorId        string `json:"successorId"`
	SerializedHash     string `json:"serializedHash"`
	SerializedSnapshot string `json:"serizalizedSnapshot"`
}

// Init
func (r *RevisionForHTTPResponse) Init(revision Revision) {
	r.Id = revision.Id
	r.SchemaName = revision.SchemaName
	r.ObjectId = revision.ObjectId
	r.SignedWithoutObjectId = revision.SignedWithoutObjectId
	r.Timestamp = revision.Timestamp
	r.AuthorizedByIndividualId = revision.AuthorizedByIndividualId
	r.AuthorizedByOtherId = revision.AuthorizedByOtherId
	r.PredecessorHash = revision.PredecessorHash
	r.PredecessorSignature = revision.PredecessorSignature
	r.ObjectData = revision.ObjectData
	r.SuccessorId = revision.SuccessorId
	r.SerializedHash = revision.SerializedHash
	r.SerializedSnapshot = revision.SerializedSnapshot
}
