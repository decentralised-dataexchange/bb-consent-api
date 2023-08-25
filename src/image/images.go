package image

import (
	"github.com/bb-consent/api/src/database"
	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// Image data type
type Image struct {
	ID   bson.ObjectId `bson:"_id,omitempty"`
	Data []byte
}

func session() *mgo.Session {
	return database.DB.Session.Copy()
}
func collection(s *mgo.Session) *mgo.Collection {
	return s.DB(database.DB.Name).C("images")
}

// Add Adds an image to image store
func Add(image []byte) (imageID string, err error) {
	s := session()
	defer s.Close()

	i := Image{bson.NewObjectId(), image}
	err = collection(s).Insert(&i)
	if err != nil {
		return "", err
	}

	return i.ID.Hex(), err
}
