package cmd

import (
	"log"
	"net/http"

	"github.com/bb-consent/api/internal/apikey"
	"github.com/bb-consent/api/internal/config"
	"github.com/bb-consent/api/internal/database"
	"github.com/bb-consent/api/internal/email"
	v2HttpPaths "github.com/bb-consent/api/internal/http_path/v2"
	"github.com/bb-consent/api/internal/iam"
	"github.com/bb-consent/api/internal/middleware"
	"github.com/bb-consent/api/internal/migrate"
	privacyDashboard "github.com/bb-consent/api/internal/privacy_dashboard"
	"github.com/bb-consent/api/internal/rbac"
	"github.com/bb-consent/api/internal/tenant"
	"github.com/bb-consent/api/internal/webhook"
	"github.com/casbin/casbin/v2"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var ConfigFileName string

func StartApiCmdHandler(cmd *cobra.Command, args []string) {

	// Load configuration
	configFile := "/opt/bb-consent/api/config/" + ConfigFileName
	loadedConfig, err := config.Load(configFile)
	if err != nil {
		log.Printf("Failed to load config file %s \n", configFile)
		panic(err)
	}

	log.Printf("config file: %s loaded\n", configFile)

	// Database
	err = database.Init(loadedConfig)
	if err != nil {
		panic(err)
	}
	log.Println("Data base session opened")

	// Webhooks
	webhook.Init(loadedConfig)
	log.Println("Webhooks configuration initialized")

	// IAM
	iam.Init(loadedConfig)
	log.Println("Iam initialized")

	// Email
	email.Init(loadedConfig)
	log.Println("Email initialized")

	// Privacy Dashboard
	privacyDashboard.Init(loadedConfig)
	log.Println("Privacy Dashboard initialized")

	// Application mode
	middleware.ApplicationModeInit(loadedConfig)
	log.Println("Application mode initialized")

	apikey.Init(loadedConfig)
	log.Println("Api key initialized")

	// Create realm and client if not exists in Keycloak
	iam.CreateRealmAndClientIfNotExists()

	// Setup Casbin auth rules
	authEnforcer, err := casbin.NewEnforcer(rbac.CreateRbacModel(), false)
	if err != nil {
		panic(err)
	}

	// Load the policy into the enforcer.
	_, err = authEnforcer.AddPolicies(rbac.GetRbacPolicies(loadedConfig.TestMode))
	if err != nil {
		panic(err)
	}

	// Execute actions based on application mode
	switch loadedConfig.ApplicationMode {
	case config.SingleTenant:
		tenant.SingleTenantConfiguration(loadedConfig)
	case config.MultiTenant:
	default:
		tenant.SingleTenantConfiguration(loadedConfig)
	}

	// Applying migration
	log.Println("Applying migrate")
	migrate.Migrate()

	// Router
	router := mux.NewRouter()
	if loadedConfig.TestMode {
		router.StrictSlash(true)
		v2HttpPaths.SetRoutes(router, authEnforcer, loadedConfig.TestMode)
	} else {
		subrouter := router.PathPrefix("/v2").Subrouter()
		v2HttpPaths.SetRoutes(subrouter, authEnforcer, loadedConfig.TestMode)
	}

	// Start server and listen in port 80
	log.Println("Listening port 80")
	http.ListenAndServe(":80", router)
}
