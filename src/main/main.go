package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/bb-consent/api/src/config"
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
	_, err := config.Load(configFile)
	if err != nil {
		log.Printf("Failed to load config file %s \n", configFile)
		panic(err)
	}

	log.Printf("config file: %s loaded\n", configFile)

	router := mux.NewRouter()
	SetRoutes(router)

	log.Println("Listening port 80")
	http.ListenAndServe(":80", router)
}
