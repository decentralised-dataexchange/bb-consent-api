package revision

import (
	"encoding/json"
	"time"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/dataagreement"
	"github.com/bb-consent/api/src/dataattribute"
	"github.com/bb-consent/api/src/policy"
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
	AuthorizedByOtherId      string `json:"authorizedByOtherId"`
	PredecessorHash          string `json:"predecessorHash"`
	PredecessorSignature     string `json:"predecessorSignature"`
	ObjectData               string `json:"objectData"`
	SuccessorId              string `json:"-"`
	SerializedHash           string `json:"-"`
	SerializedSnapshot       string `json:"-"`
}

// Init
func (r *Revision) Init(objectId string, authorisedByOtherId string, schemaName string) {
	r.Id = primitive.NewObjectID().Hex()
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
	// Update successor for previous revision
	previousRevision.updateSuccessorId(r.Id)

	// Predecessor hash
	r.updatePredecessorHash(previousRevision.SerializedHash)

	// Predecessor signature
	// TODO: Add signature for predecessor hash
	signature := ""
	r.updatePredecessorSignature(signature)

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
	err := revision.UpdateRevision(previousRevision, objectData)

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

type dataAgreementForObjectData struct {
	Id                      string                  `json:"id"`
	Version                 string                  `json:"version"`
	ControllerId            string                  `json:"controllerId"`
	ControllerUrl           string                  `json:"controllerUrl" valid:"required"`
	ControllerName          string                  `json:"controllerName" valid:"required"`
	Policy                  policy.Policy           `json:"policy" valid:"required"`
	Purpose                 string                  `json:"purpose" valid:"required"`
	PurposeDescription      string                  `json:"purposeDescription" valid:"required"`
	LawfulBasis             int                     `json:"lawfulBasis" valid:"required"`
	MethodOfUse             string                  `json:"methodOfUse" valid:"required"`
	DpiaDate                string                  `json:"dpiaDate"`
	DpiaSummaryUrl          string                  `json:"dpiaSummaryUrl"`
	Signature               dataagreement.Signature `json:"signature"`
	Active                  bool                    `json:"active"`
	Forgettable             bool                    `json:"forgettable"`
	CompatibleWithVersionId string                  `json:"compatibleWithVersionId"`
	Lifecycle               string                  `json:"lifecycle" valid:"required"`
	OrganisationId          string                  `json:"-"`
	IsDeleted               bool                    `json:"-"`
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
	}

	// Create revision
	revision := Revision{}
	revision.Init(objectData.Id, orgAdminId, config.DataAgreement)
	err := revision.CreateRevision(objectData)

	return revision, err
}

// UpdateRevisionForDataAgreement
func UpdateRevisionForDataAgreement(updatedDataAgreement dataagreement.DataAgreement, previousRevision *Revision, orgAdminId string) (Revision, error) {
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
	}

	// Update revision
	revision := Revision{}
	revision.Init(objectData.Id, orgAdminId, config.DataAgreement)
	err := revision.UpdateRevision(previousRevision, objectData)

	return revision, err
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

type dataAttributeForObjectData struct {
	Id           string   `json:"id"`
	Version      string   `json:"version"`
	AgreementIds []string `json:"agreementIds"`
	Name         string   `json:"name" valid:"required"`
	Description  string   `json:"description" valid:"required"`
	Sensitivity  bool     `json:"sensitivity"`
	Category     string   `json:"category"`
}

// CreateRevisionForDataAttribute
func CreateRevisionForDataAttribute(newDataAttribute dataattribute.DataAttribute, orgAdminId string) (Revision, error) {
	// Object data
	objectData := dataAttributeForObjectData{
		Id:           newDataAttribute.Id,
		Version:      newDataAttribute.Version,
		AgreementIds: newDataAttribute.AgreementIds,
		Name:         newDataAttribute.Name,
		Description:  newDataAttribute.Description,
		Sensitivity:  newDataAttribute.Sensitivity,
		Category:     newDataAttribute.Category,
	}

	// Create revision
	revision := Revision{}
	revision.Init(objectData.Id, orgAdminId, config.DataAttribute)
	err := revision.CreateRevision(objectData)

	return revision, err
}

// UpdateRevisionForDataAttribute
func UpdateRevisionForDataAttribute(updatedDataAttribute dataattribute.DataAttribute, previousRevision *Revision, orgAdminId string) (Revision, error) {
	// Object data
	objectData := dataAttributeForObjectData{
		Id:           updatedDataAttribute.Id,
		Version:      updatedDataAttribute.Version,
		AgreementIds: updatedDataAttribute.AgreementIds,
		Name:         updatedDataAttribute.Name,
		Description:  updatedDataAttribute.Description,
		Sensitivity:  updatedDataAttribute.Sensitivity,
		Category:     updatedDataAttribute.Category,
	}

	// Update revision
	revision := Revision{}
	revision.Init(objectData.Id, orgAdminId, config.DataAttribute)
	err := revision.UpdateRevision(previousRevision, objectData)

	return revision, err
}

func RecreateDataAttributeFromRevision(revision Revision) (dataattribute.DataAttribute, error) {

	// Deserialise revision snapshot
	var r Revision
	err := json.Unmarshal([]byte(revision.SerializedSnapshot), &r)
	if err != nil {
		return dataattribute.DataAttribute{}, err
	}

	// Deserialise data attribute
	var da dataattribute.DataAttribute
	err = json.Unmarshal([]byte(r.ObjectData), &da)
	if err != nil {
		return dataattribute.DataAttribute{}, err
	}

	return da, nil
}
