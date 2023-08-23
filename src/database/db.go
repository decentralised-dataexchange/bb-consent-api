package database

import (
	"log"
	"time"

	"github.com/bb-consent/api/src/config"
	mgo "github.com/globalsign/mgo"
)

type db struct {
	Session *mgo.Session
	Name    string
}

// DB Database session pointer
var DB db

// Init Connects to the DB, initializes the collection
func Init(config *config.Configuration) error {
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    config.DataBase.Hosts,
		Timeout:  60 * time.Second,
		Database: config.DataBase.Name,
		Username: config.DataBase.UserName,
		Password: config.DataBase.Password,
	}

	// Create a session which maintains a pool of socket connections to our MongoDB.
	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		return err
	}

	DB = db{session, config.DataBase.Name}

	err = initCollection("organizations", []string{"name"}, true)
	if err != nil {
		return err
	}

	err = initCollection("users", []string{"_id", "email", "phone"}, true)
	if err != nil {
		return err
	}

	err = initCollection("orgTypes", []string{"type"}, true)
	if err != nil {
		return err
	}

	err = initCollection("consents", []string{"userid", "orgid"}, false)
	if err != nil {
		return err
	}

	//TODO: Set unique to false
	err = initCollection("images", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("otps", []string{"phone"}, true)
	if err != nil {
		return err
	}
	err = initCollection("notifications", []string{"userid"}, false)
	if err != nil {
		return err
	}

	err = initCollection("consentHistory", []string{"userid"}, false)
	if err != nil {
		return err
	}

	err = initCollection("misc", []string{"type"}, false)
	if err != nil {
		return err
	}

	err = initCollection("actionLogs", []string{"orgid", "type"}, false)
	if err != nil {
		return err
	}

	err = initCollection("hlcHack", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("userDataRequests", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("qrcodes", []string{"orgid", "purposeid"}, true)
	if err != nil {
		return err
	}

	err = initCollection("webhooks", []string{"id"}, true)
	if err != nil {
		return err
	}

	err = initCollection("webhookDeliveries", []string{"id"}, true)
	if err != nil {
		return err
	}

	return nil
}

func initCollection(collectionName string, keys []string, unique bool) error {
	c := DB.Session.DB(DB.Name).C(collectionName)

	index := mgo.Index{
		Key:        keys,
		Unique:     unique,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err := c.EnsureIndex(index)
	if err != nil {
		return err
	}

	log.Printf("initialized %v collection", collectionName)
	return nil
}
