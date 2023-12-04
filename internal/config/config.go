package config

import (
	"encoding/json"
	"os"
)

// Iam Holds the IAM config details.
type Iam struct {
	URL           string
	Realm         string
	ClientId      string
	AdminUser     string
	AdminPassword string
	Timeout       int
}

// Twilio Twiolio account details
type Twilio struct {
	AccountSid string
	AuthToken  string
}

// SmtpConfig Smtp server details
type SmtpConfig struct {
	Username   string
	Password   string
	Host       string
	Port       int
	AdminEmail string
}

// WebhooksConfig webhooks configuration (kafka broker cluster, topic e.t.c)
type WebhooksConfig struct {
	Events []string `json:"events"`
}

// Organization organization data type
type Organization struct {
	Name        string `valid:"required"`
	Location    string `valid:"required"`
	Description string
	EulaURL     string
}

type User struct {
	Username string `valid:"required"`
	Password string `valid:"required"`
}

type PrivacyDashboard struct {
	Hostname string
	Version  string
}
type GlobalPolicy struct {
	Name                  string `json:"name"`
	Url                   string `json:"url"`
	IndustrySector        string `json:"industrySector"`
	GeographicRestriction string `json:"geographicRestriction"`
	StorageLocation       string `json:"storageLocation"`
}

// Configuration data type
type Configuration struct {
	DataBase struct {
		Hosts    []string
		Name     string
		UserName string
		Password string
	}
	ApplicationMode            string
	TestMode                   bool
	Organization               Organization
	User                       User
	ApiSecretKey               string
	Iam                        Iam
	PrivacyDashboardDeployment PrivacyDashboard
	Smtp                       SmtpConfig
	Webhooks                   WebhooksConfig
	Policy                     GlobalPolicy
}

// Load the config file
func Load(filename string) (*Configuration, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(file)
	config := &Configuration{}

	return config, decoder.Decode(&config)
}
