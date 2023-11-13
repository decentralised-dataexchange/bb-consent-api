package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/bb-consent/api/internal/common"
	"github.com/bb-consent/api/internal/config"
	daRecord "github.com/bb-consent/api/internal/dataagreement_record"
	"github.com/bb-consent/api/internal/revision"
	"github.com/bb-consent/api/internal/signature"
	"github.com/gorilla/mux"
)

// createSignatureFromUpdateSignatureRequestBody
func createSignatureFromUpdateSignatureRequestBody(toBeUpdatedSignatureObject signature.Signature, signatureReq updateSignatureforDataAgreementRecordReq) signature.Signature {

	toBeUpdatedSignatureObject.Payload = signatureReq.Signature.Payload
	toBeUpdatedSignatureObject.Signature = signatureReq.Signature.Signature
	toBeUpdatedSignatureObject.VerificationMethod = signatureReq.Signature.VerificationMethod
	toBeUpdatedSignatureObject.VerificationPayload = signatureReq.Signature.VerificationPayload
	toBeUpdatedSignatureObject.VerificationPayloadHash = signatureReq.Signature.VerificationPayloadHash
	toBeUpdatedSignatureObject.VerificationArtifact = signatureReq.Signature.VerificationArtifact
	toBeUpdatedSignatureObject.VerificationSignedBy = signatureReq.Signature.VerificationSignedBy
	toBeUpdatedSignatureObject.VerificationSignedAs = signatureReq.Signature.VerificationSignedAs
	toBeUpdatedSignatureObject.VerificationJwsHeader = signatureReq.Signature.VerificationJwsHeader
	toBeUpdatedSignatureObject.Timestamp = signatureReq.Signature.Timestamp
	toBeUpdatedSignatureObject.SignedWithoutObjectReference = signatureReq.Signature.SignedWithoutObjectReference
	toBeUpdatedSignatureObject.ObjectType = signatureReq.Signature.ObjectType
	toBeUpdatedSignatureObject.ObjectReference = signatureReq.Signature.ObjectReference

	return toBeUpdatedSignatureObject
}

type updateSignatureforDataAgreementRecordReq struct {
	Signature signature.Signature `json:"signature" valid:"required"`
}

type updateSignatureforDataAgreementRecordResp struct {
	Signature signature.Signature `json:"signature"`
}

func ServiceUpdateSignatureObject(w http.ResponseWriter, r *http.Request) {

	// Headers
	organisationId := common.Sanitize(r.Header.Get(config.OrganizationId))
	individualId := common.Sanitize(r.Header.Get(config.IndividualHeaderKey))

	dataAgreementRecordId := common.Sanitize(mux.Vars(r)[config.DataAgreementRecordId])

	// Request body
	var signatureReq updateSignatureforDataAgreementRecordReq
	b, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	json.Unmarshal(b, &signatureReq)

	// validating request payload
	valid, err := govalidator.ValidateStruct(signatureReq)
	if !valid {
		m := fmt.Sprintf("Failed to validate request body: %v", dataAgreementRecordId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Repository
	darRepo := daRecord.DataAgreementRecordRepository{}
	darRepo.Init(organisationId)

	toBeUpdatedDaRecord, err := darRepo.Get(dataAgreementRecordId)
	if err != nil {
		m := "Failed to fetch data agreement record"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	currentDataAgreementRevision, err := revision.GetLatestByObjectId(toBeUpdatedDaRecord.DataAgreementId)
	if err != nil {
		m := fmt.Sprintf("Failed to fetch latest revision for data agreement: %v", toBeUpdatedDaRecord.DataAgreementId)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	toBeUpdatedSignatureObject, err := signature.Get(toBeUpdatedDaRecord.SignatureId)
	if err != nil {
		m := "Failed to fetch signature for data agreement record"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// update signaute for data agreement record
	toBeUpdatedSignatureObject = createSignatureFromUpdateSignatureRequestBody(toBeUpdatedSignatureObject, signatureReq)

	savedSignature, err := signature.Update(toBeUpdatedSignatureObject)
	if err != nil {
		m := "Failed to update signature for data agreement record"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// update the data agreement record state
	toBeUpdatedDaRecord.State = config.Signed

	// Save data agreement to db
	_, err = darRepo.Update(toBeUpdatedDaRecord)
	if err != nil {
		m := "Failed to update data agreement record"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Create new revision
	newRevision, err := revision.UpdateRevisionForDataAgreementRecord(toBeUpdatedDaRecord, individualId, currentDataAgreementRevision)
	if err != nil {
		m := "Failed to create revision for new data agreement record"
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	// Save the revision to db
	_, err = revision.Add(newRevision)
	if err != nil {
		m := fmt.Sprintf("Failed to create new revision: %v", newRevision.Id)
		common.HandleErrorV2(w, http.StatusInternalServerError, m, err)
		return
	}

	resp := updateSignatureforDataAgreementRecordResp{
		Signature: savedSignature,
	}

	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)

}
