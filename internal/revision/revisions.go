package revision

import (
	"encoding/json"
	"strings"
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
	Id                       string `json:"id" bson:"_id,omitempty"`
	SchemaName               string `json:"schemaName"`
	ObjectId                 string `json:"objectId"`
	SignedWithoutObjectId    bool   `json:"signedWithoutObjectId"`
	Timestamp                string `json:"timestamp"`
	AuthorizedByIndividualId string `json:"authorizedByIndividualId"`
	AuthorizedByOther        string `json:"authorizedByOther"`
	PredecessorHash          string `json:"predecessorHash"`
	PredecessorSignature     string `json:"predecessorSignature"`
	ObjectData               string `json:"objectData"`
	SuccessorId              string `json:"successorId"`
	SerializedHash           string `json:"serializedHash"`
	SerializedSnapshot       string `json:"serializedSnapshot"`
}

type RevisionForSerializedSnapshot struct {
	SchemaName               string `json:"schemaName"`
	ObjectId                 string `json:"objectId"`
	SignedWithoutObjectId    bool   `json:"signedWithoutObjectId"`
	Timestamp                string `json:"timestamp"`
	AuthorizedByIndividualId string `json:"authorizedByIndividualId"`
	AuthorizedByOther        string `json:"authorizedByOther"`
	ObjectData               string `json:"objectData"`
}

// Init
func (r *Revision) Init(objectId string, authorisedByOther string, schemaName string) {
	r.Id = primitive.NewObjectID().Hex()
	r.SchemaName = schemaName
	r.ObjectId = objectId
	r.SignedWithoutObjectId = false
	r.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	r.AuthorizedByIndividualId = ""
	r.AuthorizedByOther = authorisedByOther
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

	var revisionForSerializedSnapshot RevisionForSerializedSnapshot
	revisionForSerializedSnapshot.SchemaName = r.SchemaName
	revisionForSerializedSnapshot.ObjectId = r.ObjectId
	revisionForSerializedSnapshot.SignedWithoutObjectId = r.SignedWithoutObjectId
	revisionForSerializedSnapshot.Timestamp = r.Timestamp
	revisionForSerializedSnapshot.AuthorizedByIndividualId = r.AuthorizedByIndividualId
	revisionForSerializedSnapshot.AuthorizedByOther = r.AuthorizedByOther
	revisionForSerializedSnapshot.ObjectData = r.ObjectData

	// Serialised snapshot
	// TODO: Use a standard json normalisation algorithm for .e.g JCS
	serialisedSnapshot, err := json.Marshal(revisionForSerializedSnapshot)
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
		previousRevision.updateSuccessorId(r.Id)

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
	Id                      string `json:"id" bson:"_id,omitempty"`
	Name                    string `json:"name" valid:"required"`
	Version                 string `json:"version"`
	Url                     string `json:"url" valid:"required"`
	Jurisdiction            string `json:"jurisdiction"`
	IndustrySector          string `json:"industrySector"`
	DataRetentionPeriodDays int    `json:"dataRetentionPeriodDays"`
	GeographicRestriction   string `json:"geographicRestriction"`
	StorageLocation         string `json:"storageLocation"`
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
	revision.Init(objectData.Id, orgAdminId, config.Policy)
	err := revision.CreateRevision(objectData)

	return revision, err
}

// UpdateRevisionForPolicy
func UpdateRevisionForPolicy(updatedPolicy policy.Policy, orgAdminId string) (Revision, error) {
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
	r := Revision{}
	r.Init(objectData.Id, orgAdminId, config.Policy)
	// Query for previous revisions
	previousRevision, err := GetLatestByObjectIdAndSchemaName(updatedPolicy.Id, config.Policy)
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
}

// Init
func (r *RevisionForHTTPResponse) Init(revision Revision) {
	if len(strings.TrimSpace(revision.Id)) < 1 {
		r.Id = ""
	} else {
		r.Id = revision.Id
	}
	r.SchemaName = revision.SchemaName
	r.ObjectId = revision.ObjectId
	r.SignedWithoutObjectId = revision.SignedWithoutObjectId
	r.Timestamp = revision.Timestamp
	r.AuthorizedByIndividualId = revision.AuthorizedByIndividualId
	r.AuthorizedByOther = revision.AuthorizedByOther
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
	DataUse                 string                        `json:"dataUse"`
	Dpia                    string                        `json:"dpia"`
	CompatibleWithVersion   string                        `json:"compatibleWithVersion"`
	Controller              dataagreement.Controller      `json:"controller"`
	DataSources             []dataagreement.DataSource    `json:"dataSources"`
}

// InitForDraftDataAgreement
func (r *Revision) InitForDraftDataAgreement(objectId string, authorisedByOtherId string, schemaName string) {
	r.Id = ""
	r.SchemaName = schemaName
	r.ObjectId = objectId
	r.SignedWithoutObjectId = false
	r.Timestamp = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	r.AuthorizedByIndividualId = ""
	r.AuthorizedByOther = authorisedByOtherId
}

// CreateRevisionForDataAgreement
func CreateRevisionForDataAgreement(newDataAgreement dataagreement.DataAgreement, orgAdminId string) (Revision, error) {
	// Object data
	objectData := dataAgreementForObjectData{
		Id:                      newDataAgreement.Id,
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
		DataUse:                 newDataAgreement.DataUse,
		Dpia:                    newDataAgreement.Dpia,
		CompatibleWithVersion:   newDataAgreement.CompatibleWithVersion,
		ControllerName:          newDataAgreement.ControllerName,
		Controller:              newDataAgreement.Controller,
		DataSources:             newDataAgreement.DataSources,
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
		Id:                      updatedDataAgreement.Id,
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
		DataUse:                 updatedDataAgreement.DataUse,
		Dpia:                    updatedDataAgreement.Dpia,
		CompatibleWithVersion:   updatedDataAgreement.CompatibleWithVersion,
		ControllerName:          updatedDataAgreement.ControllerName,
		Controller:              updatedDataAgreement.Controller,
		DataSources:             updatedDataAgreement.DataSources,
	}

	// Initialise revision
	r := Revision{}
	r.Init(objectData.Id, orgAdminId, config.DataAgreement)

	// Query for previous revisions
	previousRevision, err := GetLatestByObjectIdAndSchemaName(updatedDataAgreement.Id, config.DataAgreement)
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
		Id:                      newDataAgreement.Id,
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
		DataUse:                 newDataAgreement.DataUse,
		Dpia:                    newDataAgreement.Dpia,
		CompatibleWithVersion:   newDataAgreement.CompatibleWithVersion,
		ControllerName:          newDataAgreement.ControllerName,
		Controller:              newDataAgreement.Controller,
		DataSources:             newDataAgreement.DataSources,
	}

	// Create revision
	revision := Revision{}
	revision.InitForDraftDataAgreement(objectData.Id, orgAdminId, config.DataAgreement)
	err := revision.CreateRevision(objectData)

	return revision, err
}

func RecreateDataAgreementFromObjectData(objectData string) (interface{}, error) {

	// Deserialise data agreement
	var da interface{}
	err := json.Unmarshal([]byte(objectData), &da)
	if err != nil {
		return nil, err
	}

	return da, nil
}

type dataAgreementRecordForObjectData struct {
	Id                        string `json:"id" bson:"_id,omitempty"`
	DataAgreementId           string `json:"dataAgreementId"`
	DataAgreementRevisionId   string `json:"dataAgreementRevisionId"`
	DataAgreementRevisionHash string `json:"dataAgreementRevisionHash"`
	IndividualId              string `json:"individualId"`
	OptIn                     bool   `json:"optIn"`
	State                     string `json:"state" valid:"required"`
	SignatureId               string `json:"signatureId"`
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
	revision.Init(objectData.Id, orgAdminId, config.DataAgreementRecord)
	err := revision.CreateRevision(objectData)

	return revision, err
}

// UpdateRevisionForDataAgreementRecord
func UpdateRevisionForDataAgreementRecord(updatedDataAgreementRecord daRecord.DataAgreementRecord, orgAdminId string, dataAgreementRevision Revision) (Revision, error) {
	// Object data
	objectData := dataAgreementRecordForObjectData{
		Id:                        updatedDataAgreementRecord.Id,
		DataAgreementId:           updatedDataAgreementRecord.DataAgreementId,
		DataAgreementRevisionId:   dataAgreementRevision.Id,
		DataAgreementRevisionHash: dataAgreementRevision.SerializedHash,
		IndividualId:              updatedDataAgreementRecord.IndividualId,
		OptIn:                     updatedDataAgreementRecord.OptIn,
		State:                     updatedDataAgreementRecord.State,
		SignatureId:               updatedDataAgreementRecord.SignatureId,
	}

	// Update revision
	revision := Revision{}
	revision.Init(objectData.Id, orgAdminId, config.DataAgreementRecord)
	// Query for previous revisions
	previousRevision, err := GetLatestByObjectIdAndSchemaName(updatedDataAgreementRecord.Id, config.DataAgreementRecord)
	if err != nil {
		return revision, err
	}

	err = revision.UpdateRevision(&previousRevision, objectData)
	if err != nil {
		return revision, err
	}
	// Save the previous revision to db
	_, err = Update(previousRevision)
	if err != nil {
		return revision, err
	}

	return revision, err
}

func RecreateConsentRecordFromObjectData(objectData string) (daRecord.DataAgreementRecord, error) {

	// Deserialise data agreement record
	var da daRecord.DataAgreementRecord
	err := json.Unmarshal([]byte(objectData), &da)
	if err != nil {
		return da, err
	}

	return da, nil
}
