package config

import (
	"encoding/json"
	"os"
)

// JSONWebKeys The jwks keys needed to validate the tokens
type JSONWebKeys struct {
	RsaRawN string
	RsaRawE string
}

// ExternalIdentityProvidersConfiguration Holds the external identity provider configurations
type ExternalIdentityProvidersConfiguration struct {
	IdentityProviderCustomerAuthenticationFlowID string
	IdentityProviderCustomerAutoLinkFlowName     string
	IamTokenEndpoint                             string
	IamAuthEndpoint                              string
}

// Iam Holds the IAM config details.
type Iam struct {
	URL                                    string
	Realm                                  string
	Jwks                                   JSONWebKeys
	AdminUser                              string
	AdminPassword                          string
	Timeout                                int
	ExternalIdentityProvidersConfiguration ExternalIdentityProvidersConfiguration
}

// Twilio Twiolio account details
type Twilio struct {
	AccountSid string
	AuthToken  string
}

// Firebase Firebase account details
type Firebase struct {
	WebApiKey          string
	DynamicLink        string
	AndroidPackageName string
	IosAppStoreId      string
	IosBundleId        string
}

// SmtpConfig Smtp server details
type SmtpConfig struct {
	Username   string
	Password   string
	Host       string
	Port       int
	AdminEmail string
}

// KafkaBrokerConfig Kafka broker configuration
type KafkaBrokerConfig struct {
	URL     string
	GroupID string
}

// KafkaConfig Kafka cluster configuration
type KafkaConfig struct {
	Broker KafkaBrokerConfig
	Topic  string
}

// WebhooksConfig webhooks configuration (kafka broker cluster, topic e.t.c)
type WebhooksConfig struct {
	KafkaConfig KafkaConfig
}

// Configuration data type
type Configuration struct {
	DataBase struct {
		Hosts    []string
		Name     string
		UserName string
		Password string
	}
	Iam      Iam
	Twilio   Twilio
	Firebase Firebase
	Smtp     SmtpConfig
	Webhooks WebhooksConfig
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
