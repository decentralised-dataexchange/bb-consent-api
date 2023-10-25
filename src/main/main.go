package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/database"
	"github.com/bb-consent/api/src/email"
	"github.com/bb-consent/api/src/firebaseUtils"
	"github.com/bb-consent/api/src/notifications"
	v1token "github.com/bb-consent/api/src/token"
	v1Handlers "github.com/bb-consent/api/src/v1/handler"
	v1HttpPaths "github.com/bb-consent/api/src/v1/http_path"
	"github.com/bb-consent/api/src/v2/apikey"
	v2HttpPaths "github.com/bb-consent/api/src/v2/http_path"
	"github.com/bb-consent/api/src/v2/iam"
	"github.com/bb-consent/api/src/v2/middleware"
	pd "github.com/bb-consent/api/src/v2/privacy_dashboard"
	"github.com/bb-consent/api/src/v2/sms"
	v2token "github.com/bb-consent/api/src/v2/token"
	v2wh "github.com/bb-consent/api/src/v2/webhook"
	"github.com/bb-consent/api/src/webhooks"
	"github.com/casbin/casbin/v2"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var configFileName string

func startAPICmdHandlerfunc(cmd *cobra.Command, args []string) {

	// Load configuration
	configFile := "/opt/bb-consent/api/config/" + configFileName
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
	webhooks.Init(loadedConfig)
	v2wh.Init(loadedConfig)
	log.Println("Webhooks configuration initialized")

	// IAM
	v1Handlers.IamInit(loadedConfig)
	iam.Init(loadedConfig)
	log.Println("Iam initialized")

	// SMS
	sms.Init(loadedConfig)
	log.Println("SMS initialized")

	// Email
	email.Init(loadedConfig)
	log.Println("Email initialized")

	// Token
	v1token.Init(loadedConfig)
	v2token.Init(loadedConfig)
	log.Println("Token initialized")

	// Notifications
	err = notifications.Init()
	if err != nil {
		panic(err)
	}

	// Firebase
	firebaseUtils.Init(loadedConfig)
	log.Println("Firebase initialized")

	// Privacy Dashboard
	pd.Init(loadedConfig)
	log.Println("Privacy Dashboard initialized")

	// Application mode
	middleware.ApplicationModeInit(loadedConfig)
	log.Println("Application mode initialized")

	apikey.Init(loadedConfig)
	log.Println("Api key initialized")
	// Setup Casbin auth rules
	authEnforcer, err := casbin.NewEnforcer("/opt/bb-consent/api/config/auth_model.conf", "/opt/bb-consent/api/config/rbac_policy.csv")
	if err != nil {
		panic(err)
	}

	// Execute actions based on application mode
	switch loadedConfig.ApplicationMode {
	case config.SingleTenant:
		SingleTenantConfiguration(loadedConfig)
	case config.MultiTenant:
	default:
		panic("Application mode is mandatory. Specify either 'single-tenant' or 'multi-tenant'.")
	}

	// Router
	router := mux.NewRouter()
	v1HttpPaths.SetRoutes(router, authEnforcer)
	v2HttpPaths.SetRoutes(router, authEnforcer)

	// Start server and listen in port 80
	log.Println("Listening port 80")
	http.ListenAndServe(":80", router)
}

func main() {

	var rootCmd = &cobra.Command{Use: "bb-consent-api"}

	// Define the "start-api" command
	var startAPICmd = &cobra.Command{
		Use:   "start-api",
		Short: "Starts the bb consent api server",
		Run:   startAPICmdHandlerfunc,
	}

	// Define the "config" flag
	startAPICmd.Flags().StringVarP(&configFileName, "config", "c", "config-development.json", "configuration file")

	// Add the "start-api" commands to the root command
	rootCmd.AddCommand(startAPICmd)

	// Execute the CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
