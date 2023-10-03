package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/database"
	"github.com/bb-consent/api/src/email"
	"github.com/bb-consent/api/src/firebaseUtils"
	"github.com/bb-consent/api/src/handler"
	"github.com/bb-consent/api/src/kafkaUtils"
	"github.com/bb-consent/api/src/middleware"
	"github.com/bb-consent/api/src/notifications"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/webhookdispatcher"
	"github.com/bb-consent/api/src/webhooks"
	"github.com/casbin/casbin/v2"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

func handleCommandLineArgs() (configFileName string) {
	fconfigFileName := flag.String("config", "config-development.json", "configuration file")
	flag.Parse()

	return *fconfigFileName
}

func main() {

	var rootCmd = &cobra.Command{Use: "bb-consent-api"}

	var configFileName string

	// Define the "start-api" command
	var startAPICmd = &cobra.Command{
		Use:   "start-api",
		Short: "Starts the bb consent api server",
		Run: func(cmd *cobra.Command, args []string) {

			configFile := "/opt/bb-consent/api/config/" + configFileName
			config, err := config.Load(configFile)
			if err != nil {
				log.Printf("Failed to load config file %s \n", configFile)
				panic(err)
			}

			log.Printf("config file: %s loaded\n", configFile)

			err = database.Init(config)
			if err != nil {
				panic(err)
			}
			log.Println("Data base session opened")

			webhooks.Init(config)
			log.Println("Webhooks configuration initialized")

			err = kafkaUtils.Init(config)
			if err != nil {
				panic(err)
			}
			log.Println("Kafka producer client initialised")

			handler.IamInit(config)
			log.Println("Iam initialized")

			email.Init(config)
			log.Println("Email initialized")

			token.Init(config)
			log.Println("Token initialized")

			err = notifications.Init()
			if err != nil {
				panic(err)
			}

			firebaseUtils.Init(config)
			log.Println("Firebase initialized")

			middleware.ApplicationModeInit(config)
			log.Println("Application mode initialized")

			// setup casbin auth rules
			authEnforcer, err := casbin.NewEnforcer("/opt/bb-consent/api/config/auth_model.conf", "/opt/bb-consent/api/config/rbac_policy.csv")
			if err != nil {
				panic(err)
			}

			router := mux.NewRouter()
			SetRoutes(router, authEnforcer)

			log.Println("Listening port 80")
			http.ListenAndServe(":80", router)
		},
	}

	// Define the "start-webhook-dispatcher" command
	var startWebhookCmd = &cobra.Command{
		Use:   "start-webhook-dispatcher",
		Short: "Starts the webhook dispatcher",
		Run: func(cmd *cobra.Command, args []string) {

			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Println("Starting webhook dispatcher")

			configFile := "/opt/bb-consent/api/config/" + configFileName

			config, err := config.Load(configFile)
			if err != nil {
				log.Printf("Failed to load config file %s \n", configFile)
				panic(err)
			}
			log.Printf("config file: %s loaded\n", configFile)

			err = database.Init(config)
			if err != nil {
				panic(err)
			}
			log.Println("Data base session opened")

			webhookdispatcher.WebhookDispatcherInit(config)
		},
	}

	// Define the "config" flag
	startAPICmd.Flags().StringVarP(&configFileName, "config", "c", "config-development.json", "configuration file")
	startWebhookCmd.Flags().StringVarP(&configFileName, "config", "c", "config-development.json", "configuration file")

	// Add the "start-api," and "start-webhook-dispatcher" commands to the root command
	rootCmd.AddCommand(startAPICmd, startWebhookCmd)

	// Execute the CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
