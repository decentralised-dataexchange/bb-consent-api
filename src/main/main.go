package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/config"
	"github.com/bb-consent/api/src/database"
	"github.com/bb-consent/api/src/email"
	"github.com/bb-consent/api/src/firebaseUtils"
	"github.com/bb-consent/api/src/handler"
	"github.com/bb-consent/api/src/kafkaUtils"
	"github.com/bb-consent/api/src/notifications"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/webhooks"
	casbin "github.com/casbin/casbin/v2"
	"github.com/gorilla/mux"
)

func handleCommandLineArgs() (configFileName string) {
	fconfigFileName := flag.String("config", "config-development.json", "configuration file")
	flag.Parse()

	return *fconfigFileName
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting bb-consent api")

	configFileName := handleCommandLineArgs()
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

	// setup casbin auth rules
	authEnforcer, err := casbin.NewEnforcer("/opt/bb-consent/api/config/auth_model.conf", "/opt/bb-consent/api/config/rbac_policy.csv")
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	SetRoutes(router, authEnforcer)

	log.Println("Listening port 80")
	http.ListenAndServe(":80", router)
}
