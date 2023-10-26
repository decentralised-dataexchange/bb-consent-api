package onboard

import (
	"encoding/json"
	"net/http"

	"github.com/bb-consent/api/src/config"
	m "github.com/bb-consent/api/src/v2/middleware"
	"github.com/bb-consent/api/src/v2/version"
)

type readStatusResp struct {
	ApplicationMode string `json:"applicationMode"`
	Version         string `json:"version"`
	Candidate       string `json:"candidate"`
	Revision        string `json:"revision"`
}

// OnboardReadStatus
func OnboardReadStatus(w http.ResponseWriter, r *http.Request) {

	version := version.VersionFill()
	resp := readStatusResp{
		ApplicationMode: m.ApplicationMode,
		Version:         version.Version,
		Candidate:       version.Candidate,
		Revision:        version.Revision,
	}
	response, _ := json.Marshal(resp)
	w.Header().Set(config.ContentTypeHeader, config.ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
