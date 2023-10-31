package privacydashboard

import "github.com/bb-consent/api/internal/config"

// DashboardDeploymentStatus id and string of status
type DashboardDeploymentStatus struct {
	ID  int
	Str string
}

const (
	DashboardNotConfigured = 0
	DashboardRequested     = 1
	DashboardDeployed      = 2
)

var PrivacyDashboard config.PrivacyDashboard

// Init Initialize the Privacy Dashboard
func Init(config *config.Configuration) {
	PrivacyDashboard = config.PrivacyDashboard
}

// Note: Dont change the ID(s) if new type is needed then add at the end

// DashboardDeploymentStatuses Array of id and string
var DashboardDeploymentStatuses = []DashboardDeploymentStatus{
	{ID: DashboardNotConfigured, Str: "Not Configured"},
	{ID: DashboardRequested, Str: "Requested"},
	{ID: DashboardDeployed, Str: "Deployed"}}
