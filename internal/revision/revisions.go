package revision

import (
	"encoding/json"
	"time"

	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/dataagreement"
	daRecord "github.com/bb-consent/api/internal/dataagreement_record"
	"github.com/bb-consent/api/internal/policy"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Revision
type Revision struct {
	Id                       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SchemaName               string             `json:"schemaName"`
	ObjectId                 string             `json:"objectId"`
	SignedWithoutObjectId    bool               `json:"signedWithoutObjectId"`
	Timestamp                string             `json:"timestamp"`
	AuthorizedByIndividualId string             `json:"authorizedByIndividualId"`
	AuthorizedByOtherId      string             `json:"authorizedByOtherId"`
	PredecessorHash          string             `json:"predecessorHash"`
	PredecessorSignature     string             `json:"predecessorSignature"`
	ObjectData               string             `json:"objectData"`
	SuccessorId              string             `json:"-"`
	SerializedHash           string             `json:"-"`
	SerializedSnapshot       string             `json:"-"`
}

// Init
func (r *Revision) Init(objectId string, authorisedByOtherId string, schemaName string) {
	r.Id = primitive.NewObjectID()
	r.SchemaName = schemaName
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
func (r *Revision) UpdateRevision(previousRevision *Revision, objectData interface{}) error {

	if previousRevision != nil {
		// Update successor for previous revision
		previousRevision.updateSuccessorId(r.Id.Hex())

		// Predecessor hash
		r.updatePredecessorHash(previousRevision.SerializedHash)

		// Predecessor signature
		// TODO: Add signature for predecessor hash
		signature := ""
		r.updatePredecessorSignature(signature)
	}

	// Create revision
	err := r.CreateRevision(objectData)
	if err != nil {
		return err
	}

	return err
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
func CreateRevisionForPolicy(newPolicy policy.Policy, orgAdminId string) (Revision, error) {
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
	revision.Init(objectData.Id.Hex(), orgAdminId, config.Policy)
	err := revision.CreateRevision(objectData)

	return revision, err
}

// UpdateRevisionForPolicy
func UpdateRevisionForPolicy(updatedPolicy policy.Policy, previousRevision *Revision, orgAdminId string) (Revision, error) {
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
	revision.Init(objectData.Id.Hex(), orgAdminId, config.Policy)
	err := revision.UpdateRevision(nil, objectData)

	return revision, err
}

func RecreatePolicyFromRevision(revision Revision) (policy.Policy, error) {

	// Deserialise revision snapshot
	var r Revision
	err := json.Unmarshal([]byte(revision.SerializedSnapshot), &r)
	if err != nil {
		return policy.Policy{}, err
	}

	// Deserialise policy
	var p policy.Policy
	err = json.Unmarshal([]byte(r.ObjectData), &p)
	if err != nil {
		return policy.Policy{}, err
	}

	return p, nil
}

// RevisionForHTTPResponse
type RevisionForHTTPResponse struct {
	Revision
	Id                 string `json:"id"`
	SuccessorId        string `json:"successorId"`
	SerializedHash     string `json:"serializedHash"`
	SerializedSnapshot string `json:"serizalizedSnapshot"`
}

// Init
func (r *RevisionForHTTPResponse) Init(revision Revision) {
	if revision.Id.IsZero() {
		r.Id = ""
	} else {
		r.Id = revision.Id.Hex()
	}
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

type dataAgreementForObjectData struct {
	Id                      string                        `json:"id"`
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
}

// InitForDraftDataAgreement
func (r *Revision) InitForDraftDataAgreement(objectId string, authorisedByOtherId string, schemaName string) {
	r.Id = primitive.NilObjectID
	r.SchemaName = schemaName
	r.ObjectId = objectId
	r.SignedWithoutObjectId = false
	r.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	r.AuthorizedByIndividualId = ""
	r.AuthorizedByOtherId = authorisedByOtherId
}

// CreateRevisionForDataAgreement
func CreateRevisionForDataAgreement(newDataAgreement dataagreement.DataAgreement, orgAdminId string) (Revision, error) {
	// Object data
	objectData := dataAgreementForObjectData{
		Id:                      newDataAgreement.Id.Hex(),
		Version:                 newDataAgreement.Version,
		ControllerId:            newDataAgreement.ControllerId,
		ControllerUrl:           newDataAgreement.ControllerUrl,
		Policy:                  newDataAgreement.Policy,
		Purpose:                 newDataAgreement.Purpose,
		PurposeDescription:      newDataAgreement.PurposeDescription,
		LawfulBasis:             newDataAgreement.LawfulBasis,
		MethodOfUse:             newDataAgreement.MethodOfUse,
		DpiaDate:                newDataAgreement.DpiaDate,
		DpiaSummaryUrl:          newDataAgreement.DpiaSummaryUrl,
		Signature:               newDataAgreement.Signature,
		Active:                  newDataAgreement.Active,
		Forgettable:             newDataAgreement.Forgettable,
		CompatibleWithVersionId: newDataAgreement.CompatibleWithVersionId,
		Lifecycle:               newDataAgreement.Lifecycle,
		DataAttributes:          newDataAgreement.DataAttributes,
	}

	// Create revision
	revision := Revision{}
	revision.Init(objectData.Id, orgAdminId, config.DataAgreement)
	err := revision.CreateRevision(objectData)

	return revision, err
}

// UpdateRevisionForDataAgreement
func UpdateRevisionForDataAgreement(updatedDataAgreement dataagreement.DataAgreement, orgAdminId string) (Revision, error) {
	// Object data
	objectData := dataAgreementForObjectData{
		Id:                      updatedDataAgreement.Id.Hex(),
		Version:                 updatedDataAgreement.Version,
		ControllerId:            updatedDataAgreement.ControllerId,
		ControllerUrl:           updatedDataAgreement.ControllerUrl,
		Policy:                  updatedDataAgreement.Policy,
		Purpose:                 updatedDataAgreement.Purpose,
		PurposeDescription:      updatedDataAgreement.PurposeDescription,
		LawfulBasis:             updatedDataAgreement.LawfulBasis,
		MethodOfUse:             updatedDataAgreement.MethodOfUse,
		DpiaDate:                updatedDataAgreement.DpiaDate,
		DpiaSummaryUrl:          updatedDataAgreement.DpiaSummaryUrl,
		Signature:               updatedDataAgreement.Signature,
		Active:                  updatedDataAgreement.Active,
		Forgettable:             updatedDataAgreement.Forgettable,
		CompatibleWithVersionId: updatedDataAgreement.CompatibleWithVersionId,
		Lifecycle:               updatedDataAgreement.Lifecycle,
		DataAttributes:          updatedDataAgreement.DataAttributes,
	}

	// Initialise revision
	r := Revision{}
	r.Init(objectData.Id, orgAdminId, config.DataAgreement)

	// Query for previous revisions
	previousRevision, err := GetLatestByDataAgreementId(updatedDataAgreement.Id.Hex())
	if err != nil {
		// Previous revision is not present
		err = r.UpdateRevision(nil, objectData)
		if err != nil {
			return r, err
		}
	} else {
		// Previous revision is present
		err = r.UpdateRevision(&previousRevision, objectData)
		if err != nil {
			return r, err
		}

		// Save the previous revision to db
		_, err = Update(previousRevision)
		if err != nil {
			return r, err
		}
	}

	// Save the new revision to db
	_, err = Add(r)
	if err != nil {
		return r, err
	}

	return r, err
}

func RecreateDataAgreementFromRevision(revision Revision) (dataagreement.DataAgreement, error) {

	// Deserialise revision snapshot
	var r Revision
	err := json.Unmarshal([]byte(revision.SerializedSnapshot), &r)
	if err != nil {
		return dataagreement.DataAgreement{}, err
	}

	// Deserialise data agreement
	var da dataagreement.DataAgreement
	err = json.Unmarshal([]byte(r.ObjectData), &da)
	if err != nil {
		return dataagreement.DataAgreement{}, err
	}

	return da, nil
}

// CreateRevisionForDraftDataAgreement
func CreateRevisionForDraftDataAgreement(newDataAgreement dataagreement.DataAgreement, orgAdminId string) (Revision, error) {
	// Object data
	objectData := dataAgreementForObjectData{
		Id:                      newDataAgreement.Id.Hex(),
		Version:                 newDataAgreement.Version,
		ControllerId:            newDataAgreement.ControllerId,
		ControllerUrl:           newDataAgreement.ControllerUrl,
		Policy:                  newDataAgreement.Policy,
		Purpose:                 newDataAgreement.Purpose,
		PurposeDescription:      newDataAgreement.PurposeDescription,
		LawfulBasis:             newDataAgreement.LawfulBasis,
		MethodOfUse:             newDataAgreement.MethodOfUse,
		DpiaDate:                newDataAgreement.DpiaDate,
		DpiaSummaryUrl:          newDataAgreement.DpiaSummaryUrl,
		Signature:               newDataAgreement.Signature,
		Active:                  newDataAgreement.Active,
		Forgettable:             newDataAgreement.Forgettable,
		CompatibleWithVersionId: newDataAgreement.CompatibleWithVersionId,
		Lifecycle:               newDataAgreement.Lifecycle,
		DataAttributes:          newDataAgreement.DataAttributes,
	}

	// Create revision
	revision := Revision{}
	revision.InitForDraftDataAgreement(objectData.Id, orgAdminId, config.DataAgreement)
	err := revision.CreateRevision(objectData)

	return revision, err
}

type dataAgreementRecordForObjectData struct {
	Id                        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	DataAgreementId           string             `json:"dataAgreementId"`
	DataAgreementRevisionId   string             `json:"dataAgreementRevisionId"`
	DataAgreementRevisionHash string             `json:"dataAgreementRevisionHash"`
	IndividualId              string             `json:"individualId"`
	OptIn                     bool               `json:"optIn"`
	State                     string             `json:"state" valid:"required"`
	SignatureId               string             `json:"signatureId"`
}

// CreateRevisionForDataAgreementRecord
func CreateRevisionForDataAgreementRecord(newDataAgreementRecord daRecord.DataAgreementRecord, orgAdminId string) (Revision, error) {
	// Object data
	objectData := dataAgreementRecordForObjectData{
		Id:                        newDataAgreementRecord.Id,
		DataAgreementId:           newDataAgreementRecord.DataAgreementId,
		DataAgreementRevisionId:   newDataAgreementRecord.DataAgreementRevisionId,
		DataAgreementRevisionHash: newDataAgreementRecord.DataAgreementRevisionHash,
		IndividualId:              newDataAgreementRecord.IndividualId,
		OptIn:                     newDataAgreementRecord.OptIn,
		State:                     newDataAgreementRecord.State,
		SignatureId:               newDataAgreementRecord.SignatureId,
	}

	// Create revision
	revision := Revision{}
	revision.Init(objectData.Id.Hex(), orgAdminId, config.DataAgreementRecord)
	err := revision.CreateRevision(objectData)

	return revision, err
}

// UpdateRevisionForDataAgreementRecord
func UpdateRevisionForDataAgreementRecord(updatedDataAgreementRecord daRecord.DataAgreementRecord, previousRevision *Revision, orgAdminId string) (Revision, error) {
	// Object data
	objectData := dataAgreementRecordForObjectData{
		Id:                        updatedDataAgreementRecord.Id,
		DataAgreementId:           updatedDataAgreementRecord.DataAgreementId,
		DataAgreementRevisionId:   updatedDataAgreementRecord.DataAgreementRevisionId,
		DataAgreementRevisionHash: updatedDataAgreementRecord.DataAgreementRevisionHash,
		IndividualId:              updatedDataAgreementRecord.IndividualId,
		OptIn:                     updatedDataAgreementRecord.OptIn,
		State:                     updatedDataAgreementRecord.State,
		SignatureId:               updatedDataAgreementRecord.SignatureId,
	}

	// Update revision
	revision := Revision{}
	revision.Init(objectData.Id.Hex(), orgAdminId, config.DataAgreementRecord)
	err := revision.UpdateRevision(previousRevision, objectData)

	return revision, err
}
