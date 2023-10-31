package privacydashboard

import (
	"encoding/json"
	"net/http"

	pd "github.com/bb-consent/api/internal/privacy_dashboard"
)

type readPrivacyDashboardResp struct {
	HostName  string `json:"hostname"`
	Version   string `json:"version"`
	Status    int    `json:"status"`
	StatusStr string `json:"statusStr"`
}

// ConfigReadPrivacyDashboard Gets the privacy dashboard related info of the organization
func ConfigReadPrivacyDashboard(w http.ResponseWriter, r *http.Request) {

	var resp readPrivacyDashboardResp
	if pd.PrivacyDashboard.Hostname == "" {
		resp = readPrivacyDashboardResp{
			HostName:  pd.PrivacyDashboard.Hostname,
			Version:   pd.PrivacyDashboard.Version,
			Status:    pd.DashboardNotConfigured,
			StatusStr: pd.DashboardDeploymentStatuses[0].Str,
		}
	} else {
		resp = readPrivacyDashboardResp{
			HostName:  pd.PrivacyDashboard.Hostname,
			Version:   pd.PrivacyDashboard.Version,
			Status:    pd.DashboardDeployed,
			StatusStr: pd.DashboardDeploymentStatuses[2].Str,
		}
	}

	response, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
