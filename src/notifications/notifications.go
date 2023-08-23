package notifications

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var app *firebase.App

// Init Initializes the firebase admin SDK
func Init() error {
	var err error
	opt := option.WithCredentialsFile("/opt/igrant/api/config/jenkins-189019-firebase-adminsdk-p48w5-0e7b33da7b.json")
	app, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
		return err
	}
	return nil
}
