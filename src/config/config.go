package config

import (
	"encoding/json"
	"os"
)

// PricingPlan data type
type PricingPlan struct {
	ID string
}

// UsageLimit data type
type UsageLimit struct {
	APICallsLimit              string `json:"api"`
	UserLimit                  string `json:"users"`
	VerificationsLimit         string `json:"verifications"`
	VerifierAppActivationLimit string `json:"verifierappactivation"`
}

// FixedPriceModel data type
type FixedPriceModel struct {
	Price      float64
	Validity   int // Number of days
	UsageLimit UsageLimit
	SEKPlan    PricingPlan
}

// TimeCommitmentDiscount data type
type TimeCommitmentDiscount struct {
	DisplayOption string  `json:"displayOption"`
	Value         float64 `json:"value"`
}

// PayPerUserModel data type
type PayPerUserModel struct {
	BasePrice               float64
	VolumeDiscount          float64
	FirstThreshold          float64
	UserCommitmentValues    []int
	TimeCommitmentDiscounts []TimeCommitmentDiscount
	UsageLimit              UsageLimit
	SEKPlan                 PricingPlan
}

// PricingModels data type
type PricingModels struct {
	FreeTrial             FixedPriceModel
	Starter               FixedPriceModel
	PayPerUser            PayPerUserModel
	FreeTrialVerifierPlan FixedPriceModel
	StarterVerifierPlan   FixedPriceModel
	PremiumVerifierPlan   FixedPriceModel
}

// StripeConfig data type
type StripeConfig struct {
	APIKey string
}

// ServiceAgreement data type
type ServiceAgreement struct {
	URL      string
	Version  string
	FileName string
}

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

// PrivacyDashboardDeploymentConfig privacy dashboard deployment configuration
type PrivacyDashboardDeploymentConfig struct {
	KubernetesClusterIP  string
	DockerRegistryAPIURL string
	GoogleProjectID      string
	DockerImageName      string
	BackendAPIBaseURL    string
}

// SSIAriesCloudAgentDeploymentConfig SSI aries cloudagent deployment configuration
type SSIAriesCloudAgentDeploymentConfig struct {
	KubernetesClusterIP       string
	DockerRegistryAPIURL      string
	GoogleProjectID           string
	DockerImageName           string
	BackendAPIBaseURL         string
	AgentBaseURL              string
	ETHNodeRPC                string
	IntermediaryETHPrivateKey string
	ContractAddress           string
	ContractABIURL            string
}

// KafkaBrokerConfig Kafka broker configuration
type KafkaBrokerConfig struct {
	URL string
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
	Iam                   Iam
	Twilio                Twilio
	Firebase              Firebase
	VerifierAppFirebase   Firebase
	DataWalletAppFirebase Firebase
	Smtp                  SmtpConfig
	Billing               struct {
		PricingModels            PricingModels
		StripeConfig             StripeConfig
		ServiceAgreement         ServiceAgreement
		VerifierServiceAgreement ServiceAgreement
	}
	PrivacyDashboardDeployment   PrivacyDashboardDeploymentConfig
	SSIAriesCloudAgentDeployment SSIAriesCloudAgentDeploymentConfig
	Webhooks                     WebhooksConfig
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
